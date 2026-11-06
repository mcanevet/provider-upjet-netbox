// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	device "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/dcim/device"
	devicerole "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/dcim/devicerole"
	devicetype "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/dcim/devicetype"
	manufacturer "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/dcim/manufacturer"
	site "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/dcim/site"
	ipaddress "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/ipam/ipaddress"
	prefix "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/ipam/prefix"
	vlan "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/ipam/vlan"
	providerconfig "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/providerconfig"
	tenant "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/tenancy/tenant"
	cluster "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/virtualization/cluster"
	clustertype "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/virtualization/clustertype"
	virtualmachine "github.com/mcanevet/provider-upjet-netbox/internal/controller/namespaced/virtualization/virtualmachine"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		device.Setup,
		devicerole.Setup,
		devicetype.Setup,
		manufacturer.Setup,
		site.Setup,
		ipaddress.Setup,
		prefix.Setup,
		vlan.Setup,
		providerconfig.Setup,
		tenant.Setup,
		cluster.Setup,
		clustertype.Setup,
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
		devicetype.SetupGated,
		manufacturer.SetupGated,
		site.SetupGated,
		ipaddress.SetupGated,
		prefix.SetupGated,
		vlan.SetupGated,
		providerconfig.SetupGated,
		tenant.SetupGated,
		cluster.SetupGated,
		clustertype.SetupGated,
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
		devicetype.SetupWebhookWithManager,
		manufacturer.SetupWebhookWithManager,
		site.SetupWebhookWithManager,
		ipaddress.SetupWebhookWithManager,
		prefix.SetupWebhookWithManager,
		vlan.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
		tenant.SetupWebhookWithManager,
		cluster.SetupWebhookWithManager,
		clustertype.SetupWebhookWithManager,
		virtualmachine.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
