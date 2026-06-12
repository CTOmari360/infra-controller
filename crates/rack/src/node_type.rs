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

use librms::protos::rack_manager as rms;
use model::rack_type::RackHardwareTopology;

/// Default RMS node type for GB200 NVIDIA switches.
pub const DEFAULT_SWITCH_NODE_TYPE: rms::NodeType = rms::NodeType::SwitchGb200Nvidia;

/// Default RMS node type for GB200 NVIDIA compute trays.
pub const DEFAULT_COMPUTE_NODE_TYPE: rms::NodeType = rms::NodeType::ComputeGb200Nvidia;

/// Default RMS node type for GB200 LiteOn power shelves.
pub const DEFAULT_POWERSHELF_NODE_TYPE: rms::NodeType = rms::NodeType::PowershelfGb200Liteon;

pub fn is_switch_node_type(node_type: rms::NodeType) -> bool {
    matches!(
        node_type,
        rms::NodeType::SwitchGb200Nvidia | rms::NodeType::SwitchGb300Nvidia
    )
}

pub fn switch_node_type_for_topology(topology: RackHardwareTopology) -> rms::NodeType {
    match topology {
        RackHardwareTopology::Gb300Nvl36r1C2g4Topology
        | RackHardwareTopology::Gb300Nvl72r1C2g4Topology => rms::NodeType::SwitchGb300Nvidia,
        _ => rms::NodeType::SwitchGb200Nvidia,
    }
}

pub fn compute_node_type_for_topology(topology: RackHardwareTopology) -> rms::NodeType {
    match topology {
        RackHardwareTopology::Gb300Nvl36r1C2g4Topology
        | RackHardwareTopology::Gb300Nvl72r1C2g4Topology => rms::NodeType::ComputeGb300Nvidia,
        _ => rms::NodeType::ComputeGb200Nvidia,
    }
}
