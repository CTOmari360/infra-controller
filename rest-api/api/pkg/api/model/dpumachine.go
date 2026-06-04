// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"

	cdbm "github.com/NVIDIA/infra-controller/rest-api/db/pkg/db/model"
	cwssaws "github.com/NVIDIA/infra-controller/rest-api/workflow-schema/schema/site-agent/workflows/v1"
)

// APIDpuNetworkConfig represents the complete network configuration for a DPU
type APIDpuNetworkConfig struct {
	Asn                          int                          `json:"asn"`
	DhcpServers                  []string                     `json:"dhcpServers"`
	VniDevice                    string                       `json:"vniDevice"`
	ManagedHostConfig            *APIManagedHostNetworkConfig `json:"managedHostConfig"`
	ManagedHostConfigVersion     string                       `json:"managedHostConfigVersion"`
	UseAdminNetwork              bool                         `json:"useAdminNetwork"`
	AdminInterface               *APIFlatInterfaceConfig      `json:"adminInterface"`
	TenantInterfaces             []APIFlatInterfaceConfig     `json:"tenantInterfaces"`
	InstanceNetworkConfigVersion *string                      `json:"instanceNetworkConfigVersion"`
	InstanceID                   *string                      `json:"instanceId"`
	NetworkVirtualizationType    *string                      `json:"networkVirtualizationType"`
	VpcVni                       *int                         `json:"vpcVni"`
	RouteServers                 []string                     `json:"routeServers"`
	RemoteID                     string                       `json:"remoteId"`
	DeprecatedDenyPrefixes       []string                     `json:"deprecatedDenyPrefixes"`
	DpuNetworkPingerType         *string                      `json:"dpuNetworkPingerType"`
	DenyPrefixes                 []string                     `json:"denyPrefixes"`
	SiteFabricPrefixes           []string                     `json:"siteFabricPrefixes"`
	VpcIsolationBehavior         string                       `json:"vpcIsolationBehavior"`
	StatefulAclsEnabled          bool                         `json:"statefulAclsEnabled"`
	EnableDhcp                   bool                         `json:"enableDhcp"`
	HostInterfaceID              *string                      `json:"hostInterfaceId"`
	MinDpuFunctioningLinks       *int                         `json:"minDpuFunctioningLinks"`
	IsPrimaryDpu                 bool                         `json:"isPrimaryDpu"`
	InternetL3Vni                *int                         `json:"internetL3Vni"`
	DatacenterAsn                int                          `json:"datacenterAsn"`
}

// FromProto populates an APIDpuNetworkConfig from its protobuf form.
func (apnnc *APIDpuNetworkConfig) FromProto(protoConfig *cwssaws.ManagedHostNetworkConfigResponse) {
	if protoConfig == nil {
		return
	}

	apnnc.Asn = int(protoConfig.Asn)
	apnnc.DhcpServers = protoConfig.DhcpServers
	apnnc.VniDevice = protoConfig.VniDevice
	apnnc.ManagedHostConfigVersion = protoConfig.ManagedHostConfigVersion
	apnnc.UseAdminNetwork = protoConfig.UseAdminNetwork

	if protoConfig.InstanceId != nil {
		instanceID := protoConfig.InstanceId.GetValue()
		apnnc.InstanceID = &instanceID
		apnnc.InstanceNetworkConfigVersion = &protoConfig.InstanceNetworkConfigVersion
	}

	if protoConfig.NetworkVirtualizationType != nil {
		nvt := protoConfig.GetNetworkVirtualizationType().String()
		apnnc.NetworkVirtualizationType = &nvt
	}
	apnnc.RouteServers = protoConfig.RouteServers
	apnnc.RemoteID = protoConfig.RemoteId
	apnnc.DeprecatedDenyPrefixes = protoConfig.DeprecatedDenyPrefixes
	apnnc.DenyPrefixes = protoConfig.DenyPrefixes
	apnnc.SiteFabricPrefixes = protoConfig.SiteFabricPrefixes
	apnnc.VpcIsolationBehavior = protoConfig.VpcIsolationBehavior.String()
	apnnc.StatefulAclsEnabled = protoConfig.StatefulAclsEnabled
	apnnc.EnableDhcp = protoConfig.EnableDhcp
	apnnc.IsPrimaryDpu = protoConfig.IsPrimaryDpu
	apnnc.DatacenterAsn = int(protoConfig.DatacenterAsn)

	if protoConfig.ManagedHostConfig != nil {
		apnnc.ManagedHostConfig = &APIManagedHostNetworkConfig{}
		apnnc.ManagedHostConfig.FromProto(protoConfig.ManagedHostConfig)
	}

	if protoConfig.AdminInterface != nil {
		apnnc.AdminInterface = &APIFlatInterfaceConfig{}
		apnnc.AdminInterface.FromProto(protoConfig.AdminInterface)
	}

	if protoConfig.TenantInterfaces != nil {
		apnnc.TenantInterfaces = make([]APIFlatInterfaceConfig, len(protoConfig.TenantInterfaces))
		for i, protoInterface := range protoConfig.TenantInterfaces {
			if protoInterface != nil {
				apnnc.TenantInterfaces[i] = APIFlatInterfaceConfig{}
				apnnc.TenantInterfaces[i].FromProto(protoInterface)
			}
		}
	}

	if protoConfig.VpcVni != nil {
		vpcVni := int(*protoConfig.VpcVni)
		apnnc.VpcVni = &vpcVni
	}

	if protoConfig.DpuNetworkPingerType != nil {
		apnnc.DpuNetworkPingerType = protoConfig.DpuNetworkPingerType
	}

	if protoConfig.HostInterfaceId != nil {
		apnnc.HostInterfaceID = protoConfig.HostInterfaceId
	}

	if protoConfig.MinDpuFunctioningLinks != nil {
		minLinks := int(*protoConfig.MinDpuFunctioningLinks)
		apnnc.MinDpuFunctioningLinks = &minLinks
	}

	if protoConfig.InternetL3Vni != nil {
		internetL3Vni := int(*protoConfig.InternetL3Vni)
		apnnc.InternetL3Vni = &internetL3Vni
	}
}

// APIManagedHostQuarantineState represents quarantine state
type APIManagedHostQuarantineState struct {
	Mode   string  `json:"mode"`
	Reason *string `json:"reason,omitempty"`
}

// FromProto populates an APIManagedHostQuarantineState from its protobuf form.
func (amhq *APIManagedHostQuarantineState) FromProto(protoQuarantineState *cwssaws.ManagedHostQuarantineState) {
	if protoQuarantineState == nil {
		return
	}
	amhq.Mode = protoQuarantineState.Mode.String()
	amhq.Reason = protoQuarantineState.Reason
}

// APIManagedHostNetworkConfig represents the managed host network configuration
type APIManagedHostNetworkConfig struct {
	LoopbackIP      string                         `json:"loopbackIp"`
	QuarantineState *APIManagedHostQuarantineState `json:"quarantineState,omitempty"`
}

// FromProto populates an APIManagedHostNetworkConfig from its protobuf form.
func (amnc *APIManagedHostNetworkConfig) FromProto(protoConfig *cwssaws.ManagedHostNetworkConfig) {
	if protoConfig == nil {
		return
	}
	amnc.LoopbackIP = protoConfig.LoopbackIp

	quarantineState := protoConfig.QuarantineState
	if quarantineState != nil {
		amnc.QuarantineState = &APIManagedHostQuarantineState{}
		amnc.QuarantineState.FromProto(quarantineState)
	}
}

// APIFlatInterfaceConfig represents a flat interface configuration
type APIFlatInterfaceConfig struct {
	FunctionType         string                                      `json:"functionType"`
	VlanID               int                                         `json:"vlanId"`
	Vni                  int                                         `json:"vni"`
	Gateway              string                                      `json:"gateway"`
	IP                   string                                      `json:"ip"`
	InterfacePrefix      string                                      `json:"interfacePrefix"`
	VirtualFunctionID    *int                                        `json:"virtualFunctionId"`
	VpcPrefixes          []string                                    `json:"vpcPrefixes"`
	Prefix               string                                      `json:"prefix"`
	Fqdn                 string                                      `json:"fqdn"`
	BootURL              *string                                     `json:"bootUrl"`
	VpcVni               int                                         `json:"vpcVni"`
	SviIP                *string                                     `json:"sviIp"`
	TenantVrfLoopbackIP  *string                                     `json:"tenantVrfLoopbackIp"`
	IsL2Segment          bool                                        `json:"isL2Segment"`
	VpcPeerPrefixes      []string                                    `json:"vpcPeerPrefixes"`
	VpcPeerVnis          []int                                       `json:"vpcPeerVnis"`
	NetworkSecurityGroup *APIFlatInterfaceNetworkSecurityGroupConfig `json:"networkSecurityGroup"`
}

// FromProto populates an APIFlatInterfaceConfig from its protobuf form.
func (afic *APIFlatInterfaceConfig) FromProto(protoConfig *cwssaws.FlatInterfaceConfig) {
	if protoConfig == nil {
		return
	}

	afic.FunctionType = protoConfig.FunctionType.String()
	afic.VlanID = int(protoConfig.VlanId)
	afic.Vni = int(protoConfig.Vni)
	afic.Gateway = protoConfig.Gateway
	afic.IP = protoConfig.Ip
	afic.InterfacePrefix = protoConfig.InterfacePrefix

	if protoConfig.VirtualFunctionId != nil {
		virtualFunctionID := int(*protoConfig.VirtualFunctionId)
		afic.VirtualFunctionID = &virtualFunctionID
	}

	afic.VpcPrefixes = protoConfig.VpcPrefixes
	afic.Prefix = protoConfig.Prefix
	afic.Fqdn = protoConfig.Fqdn

	afic.BootURL = protoConfig.Booturl
	afic.VpcVni = int(protoConfig.VpcVni)
	afic.SviIP = protoConfig.SviIp
	afic.TenantVrfLoopbackIP = protoConfig.TenantVrfLoopbackIp
	afic.IsL2Segment = protoConfig.IsL2Segment
	afic.VpcPeerPrefixes = protoConfig.VpcPeerPrefixes

	afic.VpcPeerVnis = make([]int, len(protoConfig.VpcPeerVnis))
	for i, vni := range protoConfig.VpcPeerVnis {
		afic.VpcPeerVnis[i] = int(vni)
	}

	afic.NetworkSecurityGroup = &APIFlatInterfaceNetworkSecurityGroupConfig{}
	afic.NetworkSecurityGroup.FromProto(protoConfig.NetworkSecurityGroup)
}

// APIFlatInterfaceNetworkSecurityGroupConfig represents network security group configuration
type APIFlatInterfaceNetworkSecurityGroupConfig struct {
	ID      string                                `json:"id"`
	Version string                                `json:"version"`
	Source  string                                `json:"source"`
	Rules   []APIResolvedNetworkSecurityGroupRule `json:"rules"`
}

// FromProto populates an APIFlatInterfaceNetworkSecurityGroupConfig from its protobuf form.
func (aficsg *APIFlatInterfaceNetworkSecurityGroupConfig) FromProto(protoConfig *cwssaws.FlatInterfaceNetworkSecurityGroupConfig) {
	if protoConfig == nil {
		return
	}

	aficsg.ID = protoConfig.Id
	aficsg.Version = protoConfig.Version
	aficsg.Source = protoConfig.Source.String()
	aficsg.Rules = make([]APIResolvedNetworkSecurityGroupRule, len(protoConfig.Rules))
	for i, protoRule := range protoConfig.Rules {
		if protoRule != nil {
			aficsg.Rules[i] = APIResolvedNetworkSecurityGroupRule{}
			aficsg.Rules[i].FromProto(protoRule)
		}
	}
}

// APIResolvedNetworkSecurityGroupRule represents a resolved network security group rule
type APIResolvedNetworkSecurityGroupRule struct {
	Rule        *APINetworkSecurityGroupRule `json:"rule"`
	SrcPrefixes []string                     `json:"srcPrefixes"`
	DstPrefixes []string                     `json:"dstPrefixes"`
}

// FromProto populates an APIResolvedNetworkSecurityGroupRule from its protobuf form.
func (arnsr *APIResolvedNetworkSecurityGroupRule) FromProto(protoRule *cwssaws.ResolvedNetworkSecurityGroupRule) {
	if protoRule == nil {
		return
	}

	arnsr.Rule = NewAPINetworkSecurityGroupRule(protoRule.Rule)

	arnsr.SrcPrefixes = protoRule.SrcPrefixes
	arnsr.DstPrefixes = protoRule.DstPrefixes
}

// APIDpuMachineSoftwareComponent represents a DPU machine software component
type APIDpuMachineSoftwareComponent struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

// FromProto populates an APIDpuMachineSoftwareComponent from its protobuf form.
func (apmsc *APIDpuMachineSoftwareComponent) FromProto(protoComponent *cwssaws.MachineInventorySoftwareComponent) {
	if protoComponent == nil {
		return
	}
	apmsc.Name = protoComponent.Name
	apmsc.Version = protoComponent.Version
	apmsc.URL = protoComponent.Url
}

// APIDpuMachineInterface represents a DPU machine interface
type APIDpuMachineInterface struct {
	ID               string     `json:"id"`
	MachineID        string     `json:"machineId"`
	SegmentID        string     `json:"segmentId"`
	Hostname         string     `json:"hostname"`
	PrimaryInterface bool       `json:"primaryInterface"`
	MacAddress       string     `json:"macAddress"`
	Address          []string   `json:"address"`
	Vendor           *string    `json:"vendor"`
	Created          *time.Time `json:"created"`
	LastDhcp         *time.Time `json:"lastDhcp"`
	IsBmc            bool       `json:"isBmc"`
}

// FromProto populates an APIDpuMachineInterface from its protobuf form.
func (admif *APIDpuMachineInterface) FromProto(protoInterface *cwssaws.MachineInterface) {
	if protoInterface == nil {
		return
	}
	admif.ID = protoInterface.GetId().GetValue()
	admif.MachineID = protoInterface.GetMachineId().GetId()
	admif.SegmentID = protoInterface.GetSegmentId().GetValue()
	admif.Hostname = protoInterface.Hostname
	admif.PrimaryInterface = protoInterface.PrimaryInterface
	admif.MacAddress = protoInterface.MacAddress
	admif.Address = protoInterface.Address
	admif.Vendor = protoInterface.Vendor

	if protoInterface.Created != nil {
		created := protoInterface.Created.AsTime()
		admif.Created = &created
	}
	if protoInterface.LastDhcp != nil {
		lastDhcp := protoInterface.LastDhcp.AsTime()
		admif.LastDhcp = &lastDhcp
	}

	if protoInterface.IsBmc != nil {
		admif.IsBmc = *protoInterface.IsBmc
	}
}

// APIDpuMachine represents a DPU machine with its complete configuration
type APIDpuMachine struct {
	ID                       string                           `json:"id"`
	InfrastructureProviderID string                           `json:"infrastructureProviderId"`
	SiteID                   string                           `json:"siteId"`
	HostMachineID            string                           `json:"hostMachineId"`
	DpuAgentVersion          string                           `json:"dpuAgentVersion"`
	BMCInfo                  *APIBMCInfo                      `json:"bmcInfo"`
	DMIData                  *APIDMIData                      `json:"dmiData"`
	Interfaces               []APIDpuMachineInterface         `json:"interfaces"`
	SoftwareComponents       []APIDpuMachineSoftwareComponent `json:"softwareComponents"`
	Health                   *APIMachineHealth                `json:"health"`
	Labels                   map[string]string                `json:"labels"`
	State                    string                           `json:"state"`
	DpuNetworkConfig         APIDpuNetworkConfig              `json:"dpuNetworkConfig"`
	LastRebooted             *time.Time                       `json:"lastRebooted"`
}

// APIDpuMachines is a collection of API DPU machines.
type APIDpuMachines []APIDpuMachine

// APIDpuMachineProtoContext carries the host Machine ID, Site ID and
// Infrastructure Provider ID that the DpuMachine proto does not include and
// that the handler supplies from the host Machine's record.
type APIDpuMachineProtoContext struct {
	// HostMachineID is the ID of the host Machine the DPU is attached to, not the DPU's own ID.
	HostMachineID            string
	SiteID                   uuid.UUID
	InfrastructureProviderID uuid.UUID
}

// FromResponseProto populates APIDpuMachines from the GetDpuMachines workflow response.
func (apds *APIDpuMachines) FromResponseProto(protoDpuMachines []*cwssaws.DpuMachine, ctx APIDpuMachineProtoContext) {
	if protoDpuMachines == nil {
		return
	}

	*apds = make(APIDpuMachines, 0, len(protoDpuMachines))
	for _, protoDpuMachine := range protoDpuMachines {
		if protoDpuMachine == nil {
			continue
		}
		apiDpuMachine := APIDpuMachine{}
		apiDpuMachine.FromProto(protoDpuMachine, ctx)
		*apds = append(*apds, apiDpuMachine)
	}
}

// FromProto populates an APIDpuMachine from its protobuf form, using ctx for
// the host Machine, Site and Infrastructure Provider IDs not carried on the proto.
func (apd *APIDpuMachine) FromProto(protoDpuMachine *cwssaws.DpuMachine, ctx APIDpuMachineProtoContext) {
	if protoDpuMachine == nil {
		return
	}

	protoMachine := protoDpuMachine.GetMachine()
	if protoMachine == nil {
		return
	}

	apd.ID = protoMachine.GetId().GetId()
	apd.InfrastructureProviderID = ctx.InfrastructureProviderID.String()
	apd.SiteID = ctx.SiteID.String()
	apd.HostMachineID = ctx.HostMachineID

	if protoMachine.DpuAgentVersion != nil {
		apd.DpuAgentVersion = *protoMachine.DpuAgentVersion
	}

	if protoMachine.BmcInfo != nil {
		apd.BMCInfo = &APIBMCInfo{}
		apd.BMCInfo.FromProto(protoMachine.BmcInfo)
	}

	if protoMachine.DiscoveryInfo != nil && protoMachine.DiscoveryInfo.DmiData != nil {
		apd.DMIData = &APIDMIData{}
		apd.DMIData.FromProto(protoMachine.DiscoveryInfo.DmiData)
	}

	if protoMachine.Interfaces != nil {
		apd.Interfaces = make([]APIDpuMachineInterface, 0, len(protoMachine.Interfaces))
		for _, protoInterface := range protoMachine.Interfaces {
			if protoInterface != nil {
				apdInterface := APIDpuMachineInterface{}
				apdInterface.FromProto(protoInterface)
				apd.Interfaces = append(apd.Interfaces, apdInterface)
			}
		}
	}

	if protoMachine.Inventory != nil {
		apd.SoftwareComponents = []APIDpuMachineSoftwareComponent{}
		for _, protoComponent := range protoMachine.Inventory.Components {
			if protoComponent == nil {
				continue
			}
			apdComponent := APIDpuMachineSoftwareComponent{}
			apdComponent.FromProto(protoComponent)
			apd.SoftwareComponents = append(apd.SoftwareComponents, apdComponent)
		}
	}

	if protoMachine.Health != nil {
		apd.Health = &APIMachineHealth{}
		apd.Health.FromProto(protoMachine.Health)
	}

	var labels cdbm.Labels
	labels.FromProto(protoMachine.GetMetadata().GetLabels())
	apd.Labels = labels

	apd.State = protoMachine.State

	if protoDpuMachine.DpuNetworkConfig != nil {
		apd.DpuNetworkConfig = APIDpuNetworkConfig{}
		apd.DpuNetworkConfig.FromProto(protoDpuMachine.DpuNetworkConfig)
	}

	if protoMachine.LastRebootTime != nil {
		lastRebooted := protoMachine.LastRebootTime.AsTime()
		apd.LastRebooted = &lastRebooted
	}
}
