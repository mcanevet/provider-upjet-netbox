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

// API groups (short, without root group), named after NetBox's apps.
const (
	groupDCIM           = "dcim"
	groupTenancy        = "tenancy"
	groupIPAM           = "ipam"
	groupVirtualization = "virtualization"
)

// groupKinds maps Terraform resource names to NetBox-app-aligned API groups
// (dcim, tenancy, ipam, virtualization). Without this, upjet's default naming
// splits on the first underscore and yields e.g. group "virtual", kind
// "Machine" for netbox_virtual_machine.
var groupKinds = map[string]groupKind{
	"netbox_device":                    {groupDCIM, "Device"},
	"netbox_device_interface":          {groupDCIM, "DeviceInterface"},
	"netbox_device_role":               {groupDCIM, "DeviceRole"},
	"netbox_device_type":               {groupDCIM, "DeviceType"},
	"netbox_manufacturer":              {groupDCIM, "Manufacturer"},
	"netbox_site":                      {groupDCIM, "Site"},
	"netbox_tenant":                    {groupTenancy, "Tenant"},
	"netbox_ip_address":                {groupIPAM, "IPAddress"},
	"netbox_prefix":                    {groupIPAM, "Prefix"},
	"netbox_vlan":                      {groupIPAM, "VLAN"},
	"netbox_cluster":                   {groupVirtualization, "Cluster"},
	"netbox_cluster_type":              {groupVirtualization, "ClusterType"},
	"netbox_virtual_machine":           {groupVirtualization, "VirtualMachine"},
	"netbox_virtual_machine_interface": {groupVirtualization, "VirtualMachineInterface"},
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
