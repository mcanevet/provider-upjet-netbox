package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"

	netboxprovider "github.com/e-breuninger/terraform-provider-netbox/provider"
)

const (
	resourcePrefix = "netbox"
	modulePath     = "github.com/mcanevet/provider-upjet-netbox"
)

// terraformProviderVersion is reported by the in-process provider: the
// e-breuninger/netbox release in go.mod. config/schema.json comes from
// TERRAFORM_PROVIDER_VERSION in the Makefile, which may lag behind when the
// releases have the same resource schemas (see there).
const terraformProviderVersion = "6.0.1"

// FrameworkProvider returns the e-breuninger/netbox provider, served
// in-process (upjet's no-fork mode) instead of through the Terraform CLI.
func FrameworkProvider() fwprovider.Provider {
	return netboxprovider.New(terraformProviderVersion)()
}

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("netbox.crossplane.io"),
		// The CLI include list defaults to every resource: empty it, so that
		// all resources are served in-process.
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
		ujconfig.WithTerraformPluginFrameworkProvider(FrameworkProvider()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
			GroupKindOverrides(),
			References(),
			LateInitOverrides(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}

// GetProviderNamespaced returns the namespaced provider configuration
func GetProviderNamespaced() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("netbox.m.crossplane.io"),
		// The CLI include list defaults to every resource: empty it, so that
		// all resources are served in-process.
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
		ujconfig.WithTerraformPluginFrameworkProvider(FrameworkProvider()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
			GroupKindOverrides(),
			References(),
			LateInitOverrides(),
		),
		ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{
			ManagedResourceNamespace: "crossplane-system",
		}))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
