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

use rpc::forge::SetRestartOvsOnAdminNetworkChangeRequest;

use super::args::Args;
use crate::errors::{CarbideCliError, CarbideCliResult};
use crate::rpc::ApiClient;

pub async fn set_restart_ovs_on_use_admin_network_change(
    opts: Args,
    api_client: &ApiClient,
) -> CarbideCliResult<()> {
    let value = if opts.enable {
        "enable"
    } else if opts.force_disable {
        "force_disable"
    } else if opts.none {
        "none"
    } else {
        return Err(CarbideCliError::GenericError(
            "Must specify one of --enable, --none, or --force-disable".into(),
        ));
    };

    let machine_id = opts.machine;
    let request = SetRestartOvsOnAdminNetworkChangeRequest {
        machine_id: Some(machine_id),
        value: value.to_string(),
    };

    api_client
        .0
        .set_restart_ovs_on_admin_network_change(request)
        .await?;

    println!("restart_ovs_on_use_admin_network_change set to '{value}' for machine {machine_id}");
    Ok(())
}
