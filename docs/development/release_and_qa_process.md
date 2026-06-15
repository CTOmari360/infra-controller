# Release and QA Process

This page describes how the NVIDIA Infra Controller (NICo) project is branched,
versioned, tested, and released. It is intended for both contributors and
operators who want to understand which version of NICo they should be running.

## Where Releases Live

- **GitHub releases:** <https://github.com/NVIDIA/infra-controller/releases>
- **Issue tracker:** <https://github.com/NVIDIA/infra-controller/issues>
- **Source:** <https://github.com/NVIDIA/infra-controller>

Every published minor and patch release is available on the GitHub releases page
above, tagged with its semver version (see [Tag Naming](#tag-naming) below).

## Branches

NICo uses just two long-lived branch types — `main` and per-minor-version
release branches — together with semver tags that distinguish prereleases,
release candidates, and final releases. The `-rc` and `-pr` suffixes are
**tag** suffixes, not branch suffixes.

| Branch              | Purpose                                  | Stability                |
|---------------------|------------------------------------------|--------------------------|
| `main`              | Ongoing development                      | No stability guarantee   |
| `releases/vX.Y`     | Stabilization and release of `vX.Y.*`    | Improves over QA window, becomes stable once a non-`-rc` tag is cut |

### `main` — Ongoing Development

All changes land on `main` first. There is **no expectation of stability** on
`main`; it is not QA tested. The only tests that gate changes to `main` are the
automated tests that run in CI. Features may be incomplete and bugs may be
present at any commit.

Use `main` if you want early access to in-progress features and you accept that
things will sometimes be broken.

### `releases/vX.Y` — Release Branches

When development for a minor version is feature-complete, a new long-lived
release branch (for example, `releases/v2.1`) is cut from `main`. This single
branch holds the entire life of that minor version:

- **During the one-month QA window**, the branch carries `vX.Y-rc` tags as
  fixes land — these are *release candidates*, not final releases.
- **Once QA signs off**, a final `vX.Y.0` tag is cut on the same branch.
- **After GA**, the branch continues to host patch releases (`vX.Y.1`,
  `vX.Y.2`, …) as they are tagged.

The branch itself never carries an `-rc` suffix — only the tags on it do.
The latest non-`-rc` tag on this branch is what most users should deploy.
See [Tag Naming](#tag-naming) below.

### Tag Naming

NICo uses [semantic versioning](https://semver.org/) of the form `vX.Y.Z`:

- `X` — major version
- `Y` — minor version
- `Z` — patch version

The following tag forms appear in the repository:

- **`vX.Y.0`** — A minor release. Published as a GitHub release from
  `releases/vX.Y`.
- **`vX.Y.Z`** (where `Z > 0`) — A patch release on top of `vX.Y.0`. Also
  published as a GitHub release from `releases/vX.Y`.
- **`vX.Y-rc`** (often suffixed further, e.g. `vX.Y-rc.1`, `vX.Y-rc.2`) —
  Applied to commits on `releases/vX.Y` during the one-month QA window to
  identify release candidates. Not a final release; *not* a branch name.
- **`vX.Y-pr`** — Applied to `main` immediately after a release branch is
  cut, to indicate that `main` is now the **prerelease** for the next minor
  version. For example, the day `releases/v2.1` is cut, `main` is tagged
  `v2.2-pr`, signaling that `main` is now pre-v2.2.

## Release Cadence

NICo follows a fixed quarterly cadence with a one-month QA window.

### Minor Releases (`X.Y.0`)

Every quarter:

1. **Code complete** (last day of December, March, June, September): a new
   release branch (e.g. `releases/v2.1`) is cut from `main`.
2. Immediately after the cut, `main` is tagged with `vX.(Y+1)-pr` to mark the
   start of the next prerelease cycle on `main`.
3. The release branch is **stabilized and QA tested for one month**. During
   this window, release-candidate tags (e.g. `v2.1-rc`, `v2.1-rc.1`, …) are
   applied to commits on the branch as QA cycles through them.
4. **Final minor release** (last day of January, April, July, October): when
   QA signs off, a `vX.Y.0` tag is cut on the same `releases/vX.Y` branch and
   published as a GitHub release.

In short: minor releases ship one month after code complete.

### Patch Releases (`X.Y.Z`)

Patch releases happen on the `releases/vX.Y` branch after the corresponding
`vX.Y.0` has shipped. Patch releases are cut as needed, but at most **once per
month**, on the last day of the month. Each patch release is published on
GitHub with a `vX.Y.Z` tag.

## Which Version Should I Use?

| Goal                                                    | What to run                                                |
|---------------------------------------------------------|------------------------------------------------------------|
| Early access to in-progress features                    | Latest `main`                                              |
| Slightly more stable, willing to help shake out bugs    | Latest `vX.Y-rc[.N]` tag on `releases/vX.Y`                |
| Most stable, production-style use                       | Latest non-`-rc`, non-`-pr` tag on `releases/vX.Y`         |

Bugs found on a tagged release (`vX.Y.Z` with no `-rc` or `-pr` suffix) are
treated with the highest priority and are tracked as **QA test escapes** —
defects that slipped past the QA window and require a follow-up fix, typically
in the next patch release.

## Support Policy

At any point in time, exactly three minor releases are visible to users, each
in a different support tier. The tiers shift forward by one slot each time a
new minor release passes QA.

| Tier            | Which release      | Bug fixes?                       | Notes                                  |
|-----------------|--------------------|----------------------------------|----------------------------------------|
| **Current**     | The newest GA minor (e.g. `v2.2`) | Yes — normal bar         | Recommended for production deployments. **No new feature work lands here** — new features land in `main` and ship in the next minor release. Small, low-risk improvements may occasionally be backported alongside bug fixes. |
| **Maintenance** | One minor back (`v2.1`)           | Yes, but at a higher bar | Critical fixes and regressions only — not a destination for new feature work |
| **End-of-Life (EOL)** | Two minors back (`v2.0`)    | No                       | Unsupported. No further releases will be cut on this branch |

> **A note on terminology.** The middle tier is called **Maintenance** in this
> document. The user-supplied draft of this policy used the word *deprecated*;
> we use *Maintenance* instead because it's the more standardized industry
> term for "still supported, but on a higher bar for changes" (e.g. Kubernetes
> and PostgreSQL community releases). *Deprecated* in most ecosystems implies
> "scheduled for removal," which is closer to what we mean by **EOL**.

### Tier Transitions

When release `vX.Y` passes QA and becomes Current:

1. The release that was Current (`vX.(Y-1)`) moves to **Maintenance**.
2. The release that was Maintenance (`vX.(Y-2)`) moves to **EOL** and stops
   receiving fixes.
3. The newly Current release (`vX.Y`) begins accepting patch releases under
   the normal bar.

Because the quarterly cadence is fixed, each minor release spends roughly
three months as Current, three months as Maintenance, and is then EOL.

### Fix Backporting

NICo uses a four-level severity scheme aligned with common industry practice
(see, for example, the
[Kubernetes patch-release criteria](https://kubernetes.io/releases/patch-releases/#cherry-pick-criteria)
and the [CVSS v3.1 severity ratings](https://nvd.nist.gov/vuln-metrics/cvss)
for security issues):

| Severity     | Definition                                                                                                                                            |
|--------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Critical** | Data loss or corruption; security vulnerability rated CVSS ≥ 9.0; complete outage of a production system; no workaround available.                    |
| **High**     | Regression from the previous minor release; security vulnerability rated CVSS 7.0–8.9; major feature unusable for a typical user; workaround exists but is impractical. |
| **Medium**   | Functional bug affecting a non-critical workflow; security vulnerability rated CVSS 4.0–6.9; reasonable workaround exists.                            |
| **Low**      | Cosmetic, documentation, log-spam, minor UX, or quality-of-life issues; CVSS < 4.0.                                                                   |

A **change** is anything that is not just a bug fix: new APIs, new fields,
new flags, new dependencies, version bumps of major dependencies, refactors,
performance improvements that are not fixing a regression, etc.

The bars below apply on top of these definitions:

- **Current — "normal bar."** Accepts Critical, High, and Medium bug fixes,
  shipped via patch releases (`vX.Y.Z`). Low-severity fixes are accepted when
  cheap and safe; they may also be deferred to the next minor. **New feature
  work does not land on Current** — features land in `main` and ship in the
  next minor release. Small, low-risk *changes* (e.g. a one-line config knob,
  a clearer error message) may occasionally land alongside fixes when the
  value clearly outweighs the destabilization risk; this is the exception,
  not the rule.
- **Maintenance — "higher bar."** Accepts **Critical and High only**.
  Medium- and Low-severity bug fixes are *not* backported, and no changes
  (in the sense above) are accepted. The intent is to keep Maintenance
  releases as boring and predictable as possible: only the things that would
  force a user to upgrade anyway get backported.
- **EOL** receives no fixes regardless of severity. Users on EOL releases
  should plan an upgrade.

When in doubt about whether a fix clears the Maintenance bar, default to
"no" and link the original fix PR in a comment so the decision is auditable.

### Upgrade and Downgrade Support

| From → To                                  | Supported? |
|--------------------------------------------|------------|
| EOL → Maintenance                          | Yes        |
| EOL → Current                              | Yes        |
| Maintenance → Current                      | Yes        |
| Any → same minor, newer patch              | Yes        |
| Any backward direction (downgrade)         | **No**     |

In other words, you may skip the Maintenance tier when upgrading from EOL
straight to Current, but you may not move backward to an older minor (or to
an older patch within the same minor). If a Current release introduces a
problem that blocks you, the supported recovery is a forward-fix in the next
patch release, not a downgrade.

Downgrade support is being tracked as a potential future capability in
[issue #2019 — *feat: Need to be able to downgrade NICo versions*](https://github.com/NVIDIA/infra-controller/issues/2019);
follow that issue for the latest state.

## QA Workflow

NICo's QA process is tracked entirely in GitHub Issues, using the
[NVIDIA Infra Controller GitHub Project](https://github.com/orgs/NVIDIA/projects/142).
Every issue carries two relevant fields:

- **`Status`** — the overall lifecycle of the issue (dev side).
- **`QA Test Status`** — the QA-side lifecycle.

### Ground Rules

- **Every issue is expected to have a `QA Test Status`.** Even issues that turn
  out to need no testing must be marked `QA Not Required` — there is no
  "unset" outcome. Today this field is set manually; there is no automation
  that initializes it on issue creation.
- **Every PR is expected to have at least one linked issue.** Use GitHub's
  `Fixes #N` / `Closes #N` / `Resolves #N` keywords, or attach the PR to the
  issue from the issue's sidebar. Code changes without a linked issue should
  not merge.
- **QA decides what testing is needed, not engineering.** Engineers should not
  pre-set `QA Not Required` or otherwise short-circuit the QA triage process.
  The QA team owns triage and scoping; engineering owns the fix and the
  test-plan dev signoff.
- **Merging a PR does not close its linked issue(s).** When a PR merges, the
  linked issue should move to `Status: Verify` with `Disposition: Item Completed`
  — meaning the code is complete and the issue is now ready for QA to test.
  Closure happens only after QA passes. This is being automated via
  [PR #2584 — *ci: complete linked issues on merged PRs*](https://github.com/NVIDIA/infra-controller/pull/2584);
  the rest of this document assumes that automation is in place.

### `QA Test Status` Values and Transitions

The `QA Test Status` field walks roughly like this:

1. **`QA to triage`** — Default starting state. QA reviews the issue and
   decides what (if any) testing is required.
2. **`QA Need Info`** — QA needs clarification from the reporter or the
   engineer before they can scope the work. Returns to triage or test design
   once answered.
3. **`QA Not Required`** — Terminal QA state. The issue still proceeds through
   the normal dev workflow (`In Progress` → `Verify | Item Completed` →
   `Closed`); QA simply does not gate it. Used for internal refactors,
   dev-only tooling, doc-only changes, etc.
4. **`QA Test Design`** — QA owns the issue and is writing the test plan. The
   `QA Engineer` field is assigned at this point.
5. **`Dev Signoff Required`** — QA has drafted a test plan and is asking the
   responsible engineer to confirm that it correctly covers the change.
6. **`Test Plan Rework Required`** — Dev pushed back on the test plan. Returns
   to `QA Test Design` for revision.
7. **`Test Plan Approved`** — Dev has signed off. The plan is ready to be
   executed once the fix lands.
8. **`QA Execution`** — QA is actively running the approved test plan,
   typically after the linked PR(s) have merged and the issue has moved to
   `Status: Verify | Item Completed`.
9. **`QA Passed`** — All tests passed. The issue can move to `Status: Closed`.
10. **`QA Failed`** — Tests failed. The issue goes back to engineering for a
    fix; after the fix is merged it returns directly to `QA Execution` (the
    test plan itself does not need to be re-designed unless the failure
    reveals a gap in the plan).

```
                  ┌──────────────────┐
new issue ───────►│   QA to triage   │ (set manually today)
                  └────────┬─────────┘
                           │
        ┌──────────────────┼─────────────────────┐
        ▼                  ▼                     ▼
 ┌──────────────┐   ┌──────────────┐    ┌─────────────────┐
 │QA Not Required│  │ QA Need Info │◄──►│ QA Test Design  │
 │  (terminal,   │  └──────────────┘    │ (assign QA Eng.)│
 │ no QA gating) │                      └────────┬────────┘
 └──────────────┘                                │
                                                 ▼
                                      ┌────────────────────┐
                                      │Dev Signoff Required│
                                      └─────────┬──────────┘
                                                │
                             ┌──────────────────┴─────────────────┐
                             ▼                                    ▼
                  ┌──────────────────────┐         ┌──────────────────────────┐
                  │  Test Plan Approved  │         │Test Plan Rework Required │
                  └─────────┬────────────┘         └─────────────┬────────────┘
                            │                                    │
                            ▼                               (back to QA Test Design)
                  ┌──────────────────┐
                  │   QA Execution   │◄────────────┐
                  └────────┬─────────┘             │
                           │                       │ (fix re-merged)
                    ┌──────┴──────┐                │
                    ▼             ▼                │
              ┌─────────┐   ┌──────────┐           │
              │QA Passed│   │QA Failed │───────────┘
              └─────────┘   └──────────┘
```

### How `QA Test Status` Relates to Issue `Status`

The two fields move semi-independently:

- While dev is still working, `Status` is `In Progress` and `QA Test Status` is
  typically somewhere in the triage/test-design/signoff portion of its track.
- When the PR merges, the linked issue's `Status` flips to
  `Verify | Item Completed` (via the automation in
  [PR #2584](https://github.com/NVIDIA/infra-controller/pull/2584)), signaling
  to QA that the fix is code-complete and ready to be exercised. `QA Test
  Status` is expected to be `Test Plan Approved` (or already in `QA Execution`)
  by this point.
- Once `QA Test Status` becomes `QA Passed`, the issue's `Status` is moved to
  `Closed` with `Disposition: Item Completed`.
- If `QA Test Status` becomes `QA Failed`, the issue typically moves back to
  `Status: In Progress` so engineering can address the failure.

### Roles

- **Engineer** — writes the fix, links the PR to the issue, reviews the QA
  test plan when asked (`Dev Signoff Required`), and addresses any
  `QA Failed` outcomes.
- **QA Engineer** (set via the `QA Engineer` field, assigned at
  `QA Test Design`) — owns triage, test plan authoring, execution, and the
  final pass/fail call.

## Backward Compatibility

Breaking changes are **not allowed** anywhere in the codebase for anything that
falls under our API guarantees.

### What Is Guaranteed to Remain Backward Compatible

- The **NICo REST API**.
- The **NICo CLI** (`nicocli`) — command names, arguments, flags, values, and
  exit codes.
- **Configuration file structures** — keys, values, filenames, and locations.
- **Environment variable names and values** consumed by NICo components.

If you depend on any of the above, you can rely on them not changing
incompatibly within and across releases.

### What Is Explicitly *Not* Guaranteed

The following are considered internal and may change without notice between
releases:

- The **gRPC API** and protobuf message contents.
- The **admin CLI** (also referred to as the *debug CLI*) — a lower-level tool
  intended for operators and developers, not end users.
- The **admin UI** (also referred to as the *debug UI*) — same audience as the
  admin CLI.
- The **Vault data model** — how secrets are laid out inside HashiCorp Vault.
- The **PostgreSQL database schema** used by NICo services. See the
  [tracking issue for database backward compatibility](https://github.com/NVIDIA/infra-controller/issues)
  on GitHub for the current state of this guarantee.
- Any other internal API contract between NICo services, or persistent data
  formats used only by NICo itself.

If you build automation that depends on any of the unguaranteed items above,
expect to update it across NICo releases.

## Glossary

A few terms used on this page that may not be obvious:

- **Code complete** — the point in the cycle at which feature work for a minor
  version stops and stabilization begins. On this date, the release branch is
  cut from `main`.
- **Release candidate (rc)** — a tagged build on a `releases/vX.Y` branch
  that is a candidate for release, pending QA sign-off. Identified by the
  `-rc` suffix on the tag (e.g. `v2.1-rc`, `v2.1-rc.2`). Note: `-rc` is a
  tag suffix only; there is no `releases/...-rc` branch.
- **Prerelease (pr)** — a build of `main` that is on its way to becoming the
  next minor release. Identified by the `-pr` suffix on a tag (e.g.
  `v2.2-pr`).
- **QA sign-off** — the formal acknowledgment from QA that a release candidate
  has passed its test plan and may be promoted to a final release.
- **QA test escape** — a defect discovered in a tagged, signed-off release that
  was not caught during the QA window. These are treated as high-priority and
  typically fixed in a subsequent patch release.
- **Semver** — [semantic versioning](https://semver.org/), the `vX.Y.Z` scheme
  used by NICo where `X` is major, `Y` is minor, and `Z` is patch.
- **Test plan** — the set of test cases QA writes for a given issue during
  `QA Test Design`. The plan is what gets dev-signed-off and then executed in
  `QA Execution`.
- **Disposition** — the GitHub project field that records *why* an issue was
  closed (e.g. `Item Completed`, `Cannot reproduce`, `Will not fix`,
  `Behaves Correctly`, `Not a bug`). Independent of `QA Test Status`.
- **Current** — the most recent GA minor release. Receives bug fixes under
  the normal bar via patch releases.
- **Maintenance** — the minor release one version behind Current. Still
  supported, but only for fixes meeting a higher bar (regressions, security
  fixes, critical blockers).
- **End-of-Life (EOL)** — the minor release two versions behind Current. No
  longer receives fixes. Users should upgrade to Maintenance or Current.
