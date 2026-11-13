// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	device "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/device"
	deviceinterface "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/deviceinterface"
	deviceinterfaceprimarymacaddress "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/deviceinterfaceprimarymacaddress"
	devicerole "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/devicerole"
	devicetype "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/devicetype"
	macaddress "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/macaddress"
	manufacturer "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/manufacturer"
	moduletype "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/moduletype"
	moduletypeprofile "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/moduletypeprofile"
	site "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/site"
	ipaddress "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/ipam/ipaddress"
	prefix "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/ipam/prefix"
	vlan "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/ipam/vlan"
	providerconfig "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/providerconfig"
	tenant "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/tenancy/tenant"
	cluster "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/virtualization/cluster"
	clustertype "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/virtualization/clustertype"
	virtualmachine "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/virtualization/virtualmachine"
	virtualmachineinterface "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/virtualization/virtualmachineinterface"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		device.Setup,
		deviceinterface.Setup,
		deviceinterfaceprimarymacaddress.Setup,
		devicerole.Setup,
		devicetype.Setup,
		macaddress.Setup,
		manufacturer.Setup,
		moduletype.Setup,
		moduletypeprofile.Setup,
		site.Setup,
		ipaddress.Setup,
		prefix.Setup,
		vlan.Setup,
		providerconfig.Setup,
		tenant.Setup,
		cluster.Setup,
		clustertype.Setup,
		virtualmachine.Setup,
		virtualmachineinterface.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		device.SetupGated,
		deviceinterface.SetupGated,
		deviceinterfaceprimarymacaddress.SetupGated,
		devicerole.SetupGated,
		devicetype.SetupGated,
		macaddress.SetupGated,
		manufacturer.SetupGated,
		moduletype.SetupGated,
		moduletypeprofile.SetupGated,
		site.SetupGated,
		ipaddress.SetupGated,
		prefix.SetupGated,
		vlan.SetupGated,
		providerconfig.SetupGated,
		tenant.SetupGated,
		cluster.SetupGated,
		clustertype.SetupGated,
		virtualmachine.SetupGated,
		virtualmachineinterface.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		device.SetupWebhookWithManager,
		deviceinterface.SetupWebhookWithManager,
		deviceinterfaceprimarymacaddress.SetupWebhookWithManager,
		devicerole.SetupWebhookWithManager,
		devicetype.SetupWebhookWithManager,
		macaddress.SetupWebhookWithManager,
		manufacturer.SetupWebhookWithManager,
		moduletype.SetupWebhookWithManager,
		moduletypeprofile.SetupWebhookWithManager,
		site.SetupWebhookWithManager,
		ipaddress.SetupWebhookWithManager,
		prefix.SetupWebhookWithManager,
		vlan.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
		tenant.SetupWebhookWithManager,
		cluster.SetupWebhookWithManager,
		clustertype.SetupWebhookWithManager,
		virtualmachine.SetupWebhookWithManager,
		virtualmachineinterface.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
