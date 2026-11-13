package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ExternalNameConfigs contains all external name configurations for this
// provider. It doubles as the generation allowlist (see ExternalNameConfigured).
//
// All of these are terraform-plugin-framework resources in
// e-breuninger/netbox v6.0.0 whose "id" is a computed Int64 NetBox object ID
// and whose ImportState parses req.ID with strconv.ParseInt. Terraform writes
// that id to state as a JSON number, which upjet only accepts with
// crossplane/upjet#792 (see the replace directive in go.mod).
var ExternalNameConfigs = map[string]config.ExternalName{
	// dcim
	"netbox_device":           config.IdentifierFromProvider,
	"netbox_device_interface": config.IdentifierFromProvider,
	// Not generated (internal/provider/interface_primary_mac_address_resource.go):
	// its id is a String, the interface's id, and it PATCHes only the
	// interface's primary_mac_address.
	"netbox_device_interface_primary_mac_address": primaryMACAddressExternalName(),
	"netbox_device_role":                          config.IdentifierFromProvider,
	"netbox_device_type":                          config.IdentifierFromProvider,
	"netbox_mac_address":                          config.IdentifierFromProvider,
	"netbox_manufacturer":                         config.IdentifierFromProvider,
	"netbox_module_type":                          config.IdentifierFromProvider,
	"netbox_module_type_profile":                  config.IdentifierFromProvider,
	"netbox_site":                                 config.IdentifierFromProvider,
	// tenancy
	"netbox_tenant": config.IdentifierFromProvider,
	// ipam
	"netbox_ip_address": config.IdentifierFromProvider,
	"netbox_prefix":     config.IdentifierFromProvider,
	"netbox_vlan":       config.IdentifierFromProvider,
	// virtualization
	"netbox_cluster":                   config.IdentifierFromProvider,
	"netbox_cluster_type":              config.IdentifierFromProvider,
	"netbox_virtual_machine":           config.IdentifierFromProvider,
	"netbox_virtual_machine_interface": config.IdentifierFromProvider,
}

// primaryMACAddressExternalName is IdentifierFromProvider for the primary
// MAC address resources. upjet reads the resource before it is created, with
// an empty id; unlike the generated resources, which report that as missing,
// their Read fails with "Invalid id" (interface_primary_mac_address_resource.go
// in e-breuninger/netbox). That one diagnostic means "not created yet", so
// that the resource gets created; any other invalid id is still an error.
func primaryMACAddressExternalName() config.ExternalName {
	e := config.IdentifierFromProvider
	e.IsNotFoundDiagnosticFn = func(diags []*tfprotov6.Diagnostic) bool {
		for _, d := range diags {
			if d.Severity == tfprotov6.DiagnosticSeverityError && d.Summary == "Invalid id" && d.Detail == `expected an interface id, got ""` {
				return true
			}
		}
		return false
	}
	return e
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
