package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// references maps Terraform resource names to their cross-resource
// references: Terraform argument -> referenced Terraform resource. Each one
// adds <field>Ref and <field>Selector fields to the CRD, resolved to the
// referenced resource's external name (its NetBox ID).
var references = map[string]map[string]string{
	// NetBox requires a virtual machine to be assigned to a site and/or a
	// cluster ("A virtual machine must be assigned to a site and/or cluster.").
	"netbox_virtual_machine": {
		"site_id": "netbox_site",
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
