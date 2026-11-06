package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// groupKind is the API group (short, without root group) and Kind of a
// generated managed resource.
type groupKind struct {
	group string
	kind  string
}

// groupKinds maps Terraform resource names to NetBox-app-aligned API groups
// (dcim, tenancy, ipam, virtualization). Without this, upjet's default naming
// splits on the first underscore and yields e.g. group "virtual", kind
// "Machine" for netbox_virtual_machine.
var groupKinds = map[string]groupKind{
	"netbox_device":          {"dcim", "Device"},
	"netbox_device_role":     {"dcim", "DeviceRole"},
	"netbox_site":            {"dcim", "Site"},
	"netbox_tenant":          {"tenancy", "Tenant"},
	"netbox_ip_address":      {"ipam", "IPAddress"},
	"netbox_prefix":          {"ipam", "Prefix"},
	"netbox_vlan":            {"ipam", "VLAN"},
	"netbox_cluster":         {"virtualization", "Cluster"},
	"netbox_virtual_machine": {"virtualization", "VirtualMachine"},
}

// GroupKindOverrides sets the short group and kind for known resources.
func GroupKindOverrides() config.ResourceOption {
	return func(r *config.Resource) {
		if gk, ok := groupKinds[r.Name]; ok {
			r.ShortGroup = gk.group
			r.Kind = gk.kind
		}
	}
}
