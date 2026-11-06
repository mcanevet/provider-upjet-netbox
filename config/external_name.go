package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
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
	"netbox_device":      config.IdentifierFromProvider,
	"netbox_device_role": config.IdentifierFromProvider,
	"netbox_site":        config.IdentifierFromProvider,
	// tenancy
	"netbox_tenant": config.IdentifierFromProvider,
	// ipam
	"netbox_ip_address": config.IdentifierFromProvider,
	"netbox_prefix":     config.IdentifierFromProvider,
	"netbox_vlan":       config.IdentifierFromProvider,
	// virtualization
	"netbox_cluster":         config.IdentifierFromProvider,
	"netbox_virtual_machine": config.IdentifierFromProvider,
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
