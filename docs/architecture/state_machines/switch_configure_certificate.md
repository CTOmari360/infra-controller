# Switch Certificate Configuration (ConfigureCertificate)

This document describes how the switch state controller configures switch TLS
certificates during the **Configuring** phase. The handler delegates device
operations to **Component Manager (CM)**, which in turn calls **Rack Manager
Service (RMS)** asynchronously and polls job status until completion.

## Goals

- Install or rotate the switch NVOS certificate as part of initial switch
  bring-up, before NVOS admin credentials are stored (`RotateOsPassword`).
- Keep RMS-specific protobuf and job semantics behind the CM `NvSwitchManager`
  abstraction so the state handler stays backend-agnostic (RMS, NSM, mock).
- Persist the async **job id** in controller state so restarts can resume polling.

## Placement in the Switch FSM

`ConfigureCertificate` is a sub-state of `SwitchControllerState::Configuring`,
before `RotateOsPassword` and before `Validating`.

```mermaid
stateDiagram-v2
    direction LR

    state Configuring {
        [*] --> ConfigureCertificate_Start
        state ConfigureCertificate {
            [*] --> Start
            Start --> WaitForComplete : job submitted
            WaitForComplete --> Start : not used (terminal only)
        }
        ConfigureCertificate --> RotateOsPassword : job Completed or skipped
        RotateOsPassword --> Validating : credentials ready
    }

    ConfigureCertificate --> Error : job Failed or CM error
```

### Sub-states (`ConfigureCertificateState`)

| Sub-state | Purpose |
|-----------|---------|
| `Start` | Resolve switch endpoint, derive `domain_name`, call CM to start RMS job. |
| `WaitForComplete { job_id }` | Poll CM → RMS for job status until terminal. |

Job status values use `ConfigureSwitchCertificateState`: `Started`,
`InProgress`, `Completed`, `Failed`.

## Domain name (`domain_name`)

The state handler does **not** read site config for a certificate path. On
`Start`, it sets:

```text
domain_name = switch.rack_id.to_string()
```

| Condition | Behavior |
|-----------|----------|
| `rack_id` is `None` | Skip certificate configuration; transition to `RotateOsPassword`. |
| Component manager not configured | Skip certificate configuration; transition to `RotateOsPassword`. |
| `bmc_mac_address` is `None` | Transition to `Error`. |
| CM returns error on start | `StateHandlerError` (handler retries on next iteration). |

RMS is expected to interpret `domain_name` as the site-local certificate
identifier (for example a vault secret name or catalog entry keyed by rack).

## Component Manager API

CM exposes two methods used by the switch configuring handler:

| Method | Input | Output |
|--------|-------|--------|
| `configure_switch_certificate` | `SwitchEndpoint`, `domain_name: &str` | `job_id: String` |
| `get_configure_switch_certificate_job_status` | `job_id: &str` | `ConfigureSwitchCertificateJobStatus { state, error }` |

`SwitchEndpoint` is built from:

- Switch BMC MAC (required)
- First associated NVOS machine interface (MAC + IP if present)
- NVOS admin credentials from the credential vault (`SwitchNvosAdmin`)

### Backend matrix

| Backend | `configure_switch_certificate` | `get_configure_switch_certificate_job_status` |
|---------|----------------------------------|-----------------------------------------------|
| **RMS** (`RmsBackend`) | Resolve RMS node identity from DB; call RMS (stub today). | Poll RMS job status (stub today). |
| **Mock** | Returns `"mock-switch-cert-job"`. | Returns `Completed`. |
| **NSM** | `InvalidArgument` (not supported). | `InvalidArgument` (not supported). |

## RMS integration

### Identity resolution (RMS backend only)

Before calling RMS, `RmsBackend`:

1. Looks up `switch.id` and `switch.rack_id` via `find_rms_identities_by_macs`.
2. Builds `rms::NewNodeInfo` from the `SwitchEndpoint` and resolved identity.
3. Passes `domain_name` and device info to RMS.

If the switch has no `rack_id` in the database, identity resolution fails and CM
returns an internal error (the state handler normally skips earlier when
`switch.rack_id` is unset).

### Planned RMS RPCs (not yet in `librms`)

The current implementation uses **stubs** in
`crates/component-manager/src/rms.rs` until these RPCs exist in
`nv-rms-client`:

| RPC | Request (conceptual) | Response (conceptual) |
|-----|----------------------|------------------------|
| `configure_switch_certificate` | Device (`NewNodeInfo`), `domain_name` | `job_id`, status |
| `get_configure_switch_certificate_job_status` | `job_id` | `ConfigureSwitchCertificateState`, optional error message |

Stub behavior today:

- Start returns job id `"stub-switch-cert-job"`.
- Status poll returns `Completed` immediately.

## Sequence diagrams

### Happy path (RMS backend)

One state-controller iteration runs `Start`; a later iteration runs
`WaitForComplete` until RMS reports completion.

```mermaid
sequenceDiagram
    autonumber
    participant SCH as Switch State Handler<br/>(configuring.rs)
    participant DB as PostgreSQL
    participant Vault as Credential Manager
    participant CM as Component Manager
    participant RMS as RmsBackend
    participant RPC as RMS (librms)

    Note over SCH: State = Configuring::<br/>ConfigureCertificate(Start)

    SCH->>DB: find machine interfaces for switch
    SCH->>Vault: get SwitchNvosAdmin credentials
    SCH->>SCH: domain_name = rack_id.to_string()
    SCH->>CM: configure_switch_certificate(endpoint, domain_name)

    CM->>RMS: NvSwitchManager::configure_switch_certificate
    RMS->>DB: find_rms_identities_by_macs(bmc_mac)
    DB-->>RMS: node_id, rack_id
    RMS->>RMS: build_switch_node_info(endpoint, identity)
    RMS->>RPC: configure_switch_certificate(device, domain_name)<br/>(stub today)
    RPC-->>RMS: job_id
    RMS-->>CM: job_id
    CM-->>SCH: job_id

    SCH->>SCH: transition to ConfigureCertificate<br/>(WaitForComplete { job_id })

    Note over SCH: Next iteration(s)

    SCH->>CM: get_configure_switch_certificate_job_status(job_id)
    CM->>RMS: NvSwitchManager::get_configure_switch_certificate_job_status
    RMS->>RPC: get_configure_switch_certificate_job_status(job_id)<br/>(stub today)
    RPC-->>RMS: Started | InProgress | Completed | Failed
    RMS-->>CM: ConfigureSwitchCertificateJobStatus
    CM-->>SCH: status

    alt status = Started or InProgress
        SCH->>SCH: wait (retry next iteration)
    else status = Completed
        SCH->>SCH: transition to RotateOsPassword
    else status = Failed
        SCH->>SCH: transition to Error
    end
```

### Skip path (no rack association)

```mermaid
sequenceDiagram
    participant SCH as Switch State Handler
    participant CM as Component Manager

    Note over SCH: ConfigureCertificate(Start), rack_id = None

    SCH->>SCH: skip certificate configuration
    Note over CM: CM not called
    SCH->>SCH: transition to RotateOsPassword
```

### Error path (job failed)

```mermaid
sequenceDiagram
    participant SCH as Switch State Handler
    participant CM as Component Manager
    participant RMS as RmsBackend
    participant RPC as RMS

    Note over SCH: ConfigureCertificate(WaitForComplete { job_id })

    SCH->>CM: get_configure_switch_certificate_job_status(job_id)
    CM->>RMS: get status
    RMS->>RPC: get_configure_switch_certificate_job_status(job_id)
    RPC-->>RMS: Failed, error_message
    RMS-->>CM: state=Failed, error
    CM-->>SCH: status

    SCH->>SCH: transition to Error(cause)
```

## Persistence

Controller state is stored in `switches.controller_state` (JSON). Example
after job submission:

```json
{
  "state": "configuring",
  "config_state": {
    "ConfigureCertificate": {
      "configure_certificate": {
        "WaitForComplete": {
          "job_id": "stub-switch-cert-job"
        }
      }
    }
  }
}
```

The job id is **only** in controller state (unlike rack firmware upgrade, which
also stores a separate `firmware_upgrade_job` row). This is sufficient for a
single-switch, single-job certificate operation.

## Implementation map

| Layer | Location |
|-------|----------|
| State types | `crates/api-model/src/switch/mod.rs` — `ConfigureCertificateState`, `ConfiguringState` |
| Job status enum | `crates/api-model/src/component_manager.rs` — `ConfigureSwitchCertificateState` |
| State handler | `crates/switch-controller/src/configuring.rs` |
| CM facade | `crates/component-manager/src/component_manager.rs` |
| CM trait | `crates/component-manager/src/nv_switch_manager.rs` |
| RMS backend | `crates/component-manager/src/rms.rs` |
| Tests | `crates/api-core/src/tests/switch_state_controller/mod.rs` |

## Testing

Integration tests cover:

- Skip when `rack_id` or component manager is absent → `RotateOsPassword`
- `Start` → `WaitForComplete` with mock CM
- `WaitForComplete` → `RotateOsPassword` on success
- `WaitForComplete` → `Error` on failed job status
- `ConfigureCertificate` (completed or skipped) → `RotateOsPassword` → `Validating`

Run with `DATABASE_URL` set (sqlx test harness), filter:
`cargo test -p carbide-api-core configure_certificate`.

## Future work

1. Replace RMS stubs with real `librms` RPCs when published.
2. Map RMS job states to `ConfigureSwitchCertificateState` explicitly (mirror
   `map_rms_firmware_job_state` pattern).
3. Decide whether NSM backend should support certificate configuration or remain
   explicitly unsupported.
4. Align `domain_name` semantics with site certificate catalog / vault naming once
   RMS contract is finalized.
