/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//! Handler for SwitchControllerState::Configuring.

use carbide_secrets::credentials::{CredentialKey, Credentials};
use carbide_uuid::switch::SwitchId;
use model::component_manager::ConfigureSwitchCertificateState;
use model::switch::{
    ConfigureCertificateState, ConfiguringState, Switch, SwitchControllerState, ValidatingState,
};
use state_controller::state_handler::{
    StateHandlerContext, StateHandlerError, StateHandlerOutcome,
};

use crate::context::SwitchStateHandlerContextObjects;
use crate::endpoint;

/// Handles the Configuring state for a switch.
pub async fn handle_configuring(
    switch_id: &SwitchId,
    state: &mut Switch,
    ctx: &mut StateHandlerContext<'_, SwitchStateHandlerContextObjects>,
) -> Result<StateHandlerOutcome<SwitchControllerState>, StateHandlerError> {
    let config_state = match &state.controller_state.value {
        SwitchControllerState::Configuring { config_state } => config_state.clone(),
        _ => unreachable!("handle_configuring called with non-Configuring state"),
    };

    match config_state {
        ConfiguringState::ConfigureCertificate {
            configure_certificate,
        } => handle_configure_certificate(switch_id, state, ctx, configure_certificate).await,
        ConfiguringState::RotateOsPassword => {
            handle_rotate_os_password(switch_id, state, ctx).await
        }
    }
}

async fn handle_rotate_os_password(
    switch_id: &SwitchId,
    state: &mut Switch,
    ctx: &mut StateHandlerContext<'_, SwitchStateHandlerContextObjects>,
) -> Result<StateHandlerOutcome<SwitchControllerState>, StateHandlerError> {
    let Some(bmc_mac_address) = state.bmc_mac_address else {
        return Ok(StateHandlerOutcome::transition(
            SwitchControllerState::Error {
                cause: "No BMC MAC address on switch".to_string(),
            },
        ));
    };

    let key = CredentialKey::SwitchNvosAdmin { bmc_mac_address };

    if let Ok(Some(Credentials::UsernamePassword { .. })) =
        ctx.services.credential_manager.get_credentials(&key).await
    {
        tracing::info!(
            "Switch {:?}: NVOS admin credentials already exist in vault for BMC MAC {}",
            switch_id,
            bmc_mac_address
        );
        return Ok(StateHandlerOutcome::transition(
            SwitchControllerState::Validating {
                validating_state: ValidatingState::ValidationComplete,
            },
        ));
    }

    let mut txn = ctx.services.db_pool.begin().await?;
    let expected_switch =
        db::expected_switch::find_by_bmc_mac_address(&mut txn, bmc_mac_address).await?;
    txn.commit().await?;

    //TODO: This logic should be replaced with the logic of rotate password
    let expected_switch = match expected_switch {
        Some(es) => es,
        None => {
            return Ok(StateHandlerOutcome::transition(
                SwitchControllerState::Error {
                    cause: format!("No expected switch found for BMC MAC {}", bmc_mac_address),
                },
            ));
        }
    };

    let (username, password) = match (expected_switch.nvos_username, expected_switch.nvos_password)
    {
        (Some(username), Some(password)) => (username, password),
        _ => {
            tracing::info!(
                "Switch {:?}: no NVOS credentials in vault or expected switch for BMC MAC {}, skipping",
                switch_id,
                bmc_mac_address
            );
            return Ok(StateHandlerOutcome::transition(
                SwitchControllerState::Validating {
                    validating_state: ValidatingState::ValidationComplete,
                },
            ));
        }
    };

    let credentials = Credentials::UsernamePassword { username, password };

    ctx.services
        .credential_manager
        .set_credentials(&key, &credentials)
        .await
        .map_err(|e| {
            StateHandlerError::GenericError(eyre::eyre!(
                "Switch {:?}: failed to store NVOS credentials in vault: {}",
                switch_id,
                e
            ))
        })?;

    tracing::info!(
        "Switch {:?}: stored NVOS admin credentials from expected switch into vault for BMC MAC {}",
        switch_id,
        bmc_mac_address
    );

    Ok(StateHandlerOutcome::transition(
        SwitchControllerState::Validating {
            validating_state: ValidatingState::ValidationComplete,
        },
    ))
}

async fn handle_configure_certificate(
    switch_id: &SwitchId,
    state: &mut Switch,
    ctx: &mut StateHandlerContext<'_, SwitchStateHandlerContextObjects>,
    configure_certificate: ConfigureCertificateState,
) -> Result<StateHandlerOutcome<SwitchControllerState>, StateHandlerError> {
    match configure_certificate {
        ConfigureCertificateState::Start => {
            handle_configure_certificate_start(switch_id, state, ctx).await
        }
        ConfigureCertificateState::WaitForComplete { job_id } => {
            handle_configure_certificate_wait_for_complete(switch_id, ctx, &job_id).await
        }
    }
}

async fn handle_configure_certificate_start(
    switch_id: &SwitchId,
    state: &Switch,
    ctx: &mut StateHandlerContext<'_, SwitchStateHandlerContextObjects>,
) -> Result<StateHandlerOutcome<SwitchControllerState>, StateHandlerError> {
    let Some(_rack_id) = state.rack_id.as_ref() else {
        tracing::info!(
            "Switch {:?}: no rack association, skipping certificate configuration",
            switch_id
        );
        return Ok(StateHandlerOutcome::transition(
            SwitchControllerState::Configuring {
                config_state: ConfiguringState::RotateOsPassword,
            },
        ));
    };
    let domain_name = "site-wide".to_string();

    let Some(component_manager) = ctx.services.component_manager.as_ref() else {
        tracing::info!(
            "Switch {:?}: component manager is not configured, skipping certificate configuration",
            switch_id
        );
        return Ok(StateHandlerOutcome::transition(
            SwitchControllerState::Configuring {
                config_state: ConfiguringState::RotateOsPassword,
            },
        ));
    };

    if state.bmc_mac_address.is_none() {
        return Ok(StateHandlerOutcome::transition(
            SwitchControllerState::Error {
                cause: "No BMC MAC address on switch".to_string(),
            },
        ));
    }

    let endpoint = endpoint::resolve_switch_endpoint(
        switch_id,
        &ctx.services.db_pool,
        &ctx.services.credential_manager,
    )
    .await?;
    let job_id = component_manager
        .configure_switch_certificate(&endpoint, &domain_name)
        .await
        .map_err(|error| {
            StateHandlerError::GenericError(eyre::eyre!(
                "Switch {:?}: failed to start switch certificate configuration: {}",
                switch_id,
                error
            ))
        })?;

    tracing::info!(
        %job_id,
        "Switch {:?}: started switch certificate configuration",
        switch_id
    );

    Ok(StateHandlerOutcome::transition(
        SwitchControllerState::Configuring {
            config_state: ConfiguringState::ConfigureCertificate {
                configure_certificate: ConfigureCertificateState::WaitForComplete { job_id },
            },
        },
    ))
}

async fn handle_configure_certificate_wait_for_complete(
    switch_id: &SwitchId,
    ctx: &mut StateHandlerContext<'_, SwitchStateHandlerContextObjects>,
    job_id: &str,
) -> Result<StateHandlerOutcome<SwitchControllerState>, StateHandlerError> {
    let Some(component_manager) = ctx.services.component_manager.as_ref() else {
        return Ok(StateHandlerOutcome::transition(
            SwitchControllerState::Error {
                cause:
                    "component manager is not configured while waiting for switch certificate job"
                        .to_string(),
            },
        ));
    };

    let status = component_manager
        .get_configure_switch_certificate_job_status(job_id)
        .await
        .map_err(|error| {
            StateHandlerError::GenericError(eyre::eyre!(
                "Switch {:?}: failed to get switch certificate job status for {}: {}",
                switch_id,
                job_id,
                error
            ))
        })?;

    match status.state {
        ConfigureSwitchCertificateState::Completed => {
            tracing::info!(
                %job_id,
                "Switch {:?}: switch certificate configuration completed",
                switch_id
            );
            Ok(StateHandlerOutcome::transition(
                SwitchControllerState::Configuring {
                    config_state: ConfiguringState::RotateOsPassword,
                },
            ))
        }
        ConfigureSwitchCertificateState::Failed => {
            let cause = status
                .error
                .unwrap_or_else(|| "switch certificate configuration failed".to_string());
            Ok(StateHandlerOutcome::transition(
                SwitchControllerState::Error { cause },
            ))
        }
        ConfigureSwitchCertificateState::Started | ConfigureSwitchCertificateState::InProgress => {
            Ok(StateHandlerOutcome::wait(format!(
                "switch certificate job {job_id} in progress"
            )))
        }
    }
}
