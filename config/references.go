package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// references maps Terraform resource names to their cross-resource
// references: Terraform argument -> referenced Terraform resource. Each one
// adds <field>Ref and <field>Selector fields to the CRD, resolved to the
// referenced resource's external name (its NetBox ID).
var references = map[string]map[string]string{
	// The arguments NetBox requires (device_type_id, role_id, site_id).
	"netbox_device": {
		"device_type_id": "netbox_device_type",
		"role_id":        "netbox_device_role",
		"site_id":        "netbox_site",
	},
	"netbox_device_interface": {
		"device_id": "netbox_device",
	},
	"netbox_device_type": {
		"manufacturer_id": "netbox_manufacturer",
	},
	"netbox_cluster": {
		"cluster_type_id": "netbox_cluster_type",
	},
	// NetBox requires a virtual machine to be assigned to a site and/or a
	// cluster ("A virtual machine must be assigned to a site and/or cluster.").
	"netbox_virtual_machine": {
		"site_id": "netbox_site",
	},
	"netbox_virtual_machine_interface": {
		"virtual_machine_id": "netbox_virtual_machine",
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
