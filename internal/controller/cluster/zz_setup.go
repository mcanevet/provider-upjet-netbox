// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	device "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/device"
	devicerole "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/devicerole"
	site "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/dcim/site"
	ipaddress "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/ipam/ipaddress"
	prefix "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/ipam/prefix"
	vlan "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/ipam/vlan"
	providerconfig "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/providerconfig"
	tenant "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/tenancy/tenant"
	cluster "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/virtualization/cluster"
	virtualmachine "github.com/mcanevet/provider-upjet-netbox/internal/controller/cluster/virtualization/virtualmachine"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		device.Setup,
		devicerole.Setup,
		site.Setup,
		ipaddress.Setup,
		prefix.Setup,
		vlan.Setup,
		providerconfig.Setup,
		tenant.Setup,
		cluster.Setup,
		virtualmachine.Setup,
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
		devicerole.SetupGated,
		site.SetupGated,
		ipaddress.SetupGated,
		prefix.SetupGated,
		vlan.SetupGated,
		providerconfig.SetupGated,
		tenant.SetupGated,
		cluster.SetupGated,
		virtualmachine.SetupGated,
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
		devicerole.SetupWebhookWithManager,
		site.SetupWebhookWithManager,
		ipaddress.SetupWebhookWithManager,
		prefix.SetupWebhookWithManager,
		vlan.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
		tenant.SetupWebhookWithManager,
		cluster.SetupWebhookWithManager,
		virtualmachine.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
