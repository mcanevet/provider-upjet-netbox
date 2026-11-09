package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Terraform resource names used as reference targets.
const (
	tfClusterType       = "netbox_cluster_type"
	tfDevice            = "netbox_device"
	tfDeviceRole        = "netbox_device_role"
	tfDeviceType        = "netbox_device_type"
	tfManufacturer      = "netbox_manufacturer"
	tfModuleType        = "netbox_module_type"
	tfModuleTypeProfile = "netbox_module_type_profile"
	tfSite              = "netbox_site"
	tfVirtualMachine    = "netbox_virtual_machine"
)

// references maps Terraform resource names to their cross-resource
// references: Terraform argument -> referenced Terraform resource. Each one
// adds <field>Ref and <field>Selector fields to the CRD, resolved to the
// referenced resource's external name (its NetBox ID).
var references = map[string]map[string]string{
	// The arguments NetBox requires (device_type_id, role_id, site_id).
	"netbox_device": {
		"device_type_id": tfDeviceType,
		"role_id":        tfDeviceRole,
		"site_id":        tfSite,
	},
	"netbox_device_interface": {
		"device_id": tfDevice,
	},
	"netbox_device_type": {
		"manufacturer_id": tfManufacturer,
	},
	"netbox_module_type": {
		"manufacturer_id": tfManufacturer,
		"profile_id":      tfModuleTypeProfile,
	},
	// A module bay template belongs to exactly one of a device type and a
	// module type.
	"netbox_module_bay_template": {
		"device_type_id": tfDeviceType,
		"module_type_id": tfModuleType,
	},
	"netbox_cluster": {
		"cluster_type_id": tfClusterType,
	},
	// NetBox requires a virtual machine to be assigned to a site and/or a
	// cluster ("A virtual machine must be assigned to a site and/or cluster.").
	"netbox_virtual_machine": {
		"site_id": tfSite,
	},
	"netbox_virtual_machine_interface": {
		"virtual_machine_id": tfVirtualMachine,
	},
}

// References configures the cross-resource references of known resources.
func References() config.ResourceOption {
	return func(r *config.Resource) {
		refs, ok := references[r.Name]
		if !ok {
			return
		}
		if r.References == nil {
			r.References = config.References{}
		}
		for field, target := range refs {
			r.References[field] = config.Reference{TerraformName: target}
		}
	}
}
