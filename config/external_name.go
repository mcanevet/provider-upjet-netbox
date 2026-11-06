package config

import (
	"encoding/json"
	"strconv"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/pkg/errors"
)

// ExternalNameConfigs contains all external name configurations for this
// provider. It doubles as the generation allowlist (see ExternalNameConfigured).
//
// All of these are terraform-plugin-framework resources in
// e-breuninger/netbox v6.0.0 whose "id" is a computed Int64 NetBox object ID
// and whose ImportState parses req.ID with strconv.ParseInt.
var ExternalNameConfigs = map[string]config.ExternalName{
	// dcim
	"netbox_device":      netboxID(),
	"netbox_device_role": netboxID(),
	"netbox_site":        netboxID(),
	// tenancy
	"netbox_tenant": netboxID(),
	// ipam
	"netbox_ip_address": netboxID(),
	"netbox_prefix":     netboxID(),
	"netbox_vlan":       netboxID(),
	// virtualization
	"netbox_cluster":         netboxID(),
	"netbox_virtual_machine": netboxID(),
}

// netboxID is IdentifierFromProvider whose GetExternalNameFn also accepts a
// numeric "id". upjet's stock config.IDAsExternalName only accepts strings,
// but the v6 framework resources declare "id" as schema.Int64Attribute, so
// Terraform writes it to state as a JSON number.
func netboxID() config.ExternalName {
	e := config.IdentifierFromProvider
	e.GetExternalNameFn = func(tfstate map[string]any) (string, error) {
		switch id := tfstate["id"].(type) {
		case string:
			if id != "" {
				return id, nil
			}
		case float64:
			return strconv.FormatInt(int64(id), 10), nil
		case int64:
			return strconv.FormatInt(id, 10), nil
		case json.Number:
			return id.String(), nil
		}
		return "", errors.New("cannot find id in tfstate")
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
