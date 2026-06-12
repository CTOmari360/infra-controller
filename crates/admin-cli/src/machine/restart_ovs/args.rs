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

use carbide_uuid::machine::MachineId;
use clap::Parser;

#[derive(Parser, Debug)]
pub struct Args {
    #[clap(long, help = "Host machine ID")]
    pub machine: MachineId,

    #[clap(
        long,
        action,
        group = "toggle",
        help = "Enable OVS restart on admin network change for this machine (overrides site config if it is unset)"
    )]
    pub enable: bool,

    #[clap(
        long,
        action,
        group = "toggle",
        help = "Set OVS restart for this machine to none (use the site config default)"
    )]
    pub none: bool,

    #[clap(
        long,
        action,
        group = "toggle",
        help = "Force-disable OVS restart for this machine (overrides site config if it is set)"
    )]
    pub force_disable: bool,
}
