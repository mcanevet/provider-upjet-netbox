package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// lateInitIgnoredFields maps Terraform resource names to the arguments
// late-initialization must not copy from the observed state into
// spec.forProvider.
var lateInitIgnoredFields = map[string][]string{
	// The assignment can be set as device_interface_id,
	// virtual_machine_interface_id or the assigned_object_type/_id pair, which
	// conflict with each other. All four are computed from the assignment, so
	// late-initializing them would turn one setting into conflicting ones.
	"netbox_mac_address": {
		"assigned_object_type",
		"assigned_object_id",
		"device_interface_id",
		"virtual_machine_interface_id",
	},
}

// LateInitOverrides configures the late-initialization of known resources.
func LateInitOverrides() config.ResourceOption {
	return func(r *config.Resource) {
		if fields, ok := lateInitIgnoredFields[r.Name]; ok {
			r.LateInitializer.IgnoredFields = append(r.LateInitializer.IgnoredFields, fields...)
		}
	}
}
