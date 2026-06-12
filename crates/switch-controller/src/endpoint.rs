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

//! Helpers for building `SwitchEndpoint` values from api-db endpoint rows.

use std::net::IpAddr;
use std::sync::Arc;

use carbide_secrets::credentials::{CredentialKey, CredentialManager, Credentials};
use carbide_uuid::switch::SwitchId;
use component_manager::nv_switch_manager::SwitchEndpoint;
use db::switch::SwitchEndpointRow;
use mac_address::MacAddress;
use state_controller::state_handler::StateHandlerError;

fn placeholder_ip() -> IpAddr {
    "0.0.0.0".parse().expect("valid placeholder IP")
}

async fn fetch_switch_nvos_credentials(
    credential_manager: &Arc<dyn CredentialManager>,
    bmc_mac: MacAddress,
) -> Credentials {
    let key = CredentialKey::SwitchNvosAdmin {
        bmc_mac_address: bmc_mac,
    };
    match credential_manager.get_credentials(&key).await {
        Ok(Some(credentials)) => credentials,
        _ => Credentials::UsernamePassword {
            username: String::new(),
            password: String::new(),
        },
    }
}

pub fn switch_endpoint_from_row(
    row: &SwitchEndpointRow,
    nvos_credentials: Credentials,
) -> SwitchEndpoint {
    let nvos_mac = row.nvos_mac.unwrap_or(row.bmc_mac);
    let nvos_ip = row.nvos_ip.unwrap_or_else(placeholder_ip);

    SwitchEndpoint {
        bmc_ip: row.bmc_ip,
        bmc_mac: row.bmc_mac,
        nvos_ip,
        nvos_mac,
        bmc_credentials: nvos_credentials.clone(),
        nvos_credentials,
    }
}

/// Resolve a switch to a CM `SwitchEndpoint` using the shared api-db query also
/// used by the component-manager gRPC handler.
pub async fn resolve_switch_endpoint(
    switch_id: &SwitchId,
    db_pool: &sqlx::PgPool,
    credential_manager: &Arc<dyn CredentialManager>,
) -> Result<SwitchEndpoint, StateHandlerError> {
    let rows = db::switch::find_switch_endpoints_by_ids(db_pool, std::slice::from_ref(switch_id))
        .await
        .map_err(StateHandlerError::from)?;
    let row = rows.into_iter().next().ok_or_else(|| {
        StateHandlerError::GenericError(eyre::eyre!(
            "Switch {:?}: no endpoint row found in database",
            switch_id
        ))
    })?;
    let nvos_credentials = fetch_switch_nvos_credentials(credential_manager, row.bmc_mac).await;
    Ok(switch_endpoint_from_row(&row, nvos_credentials))
}
