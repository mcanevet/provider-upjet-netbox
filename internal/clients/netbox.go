package clients

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/upjet/v2/pkg/terraform"

	clusterv1beta1 "github.com/mcanevet/provider-upjet-netbox/apis/cluster/v1beta1"
	namespacedv1beta1 "github.com/mcanevet/provider-upjet-netbox/apis/namespaced/v1beta1"
	"github.com/mcanevet/provider-upjet-netbox/config"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal netbox credentials as JSON"

	// provider configuration keys
	keyServerURL      = "server_url"
	keyAPIToken       = "api_token"
	keyRequestTimeout = "request_timeout"
	keySkipVersion    = "skip_version_check"

	// defaultRequestTimeout is the NetBox API request timeout, in seconds,
	// used when the credentials do not set request_timeout. The Terraform
	// provider defaults to 10s (provider_gen.go in terraform-provider-netbox
	// v6.0.0), which is too short: when a create outlasts it, NetBox still
	// creates the object but Terraform records nothing, so the object is
	// orphaned and every retry fails on NetBox's uniqueness constraints.
	defaultRequestTimeout = 60
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		pcSpec, err := resolveProviderConfig(ctx, client, mg)
		if err != nil {
			return terraform.Setup{}, errors.Wrap(err, "cannot resolve provider config")
		}

		data, err := resource.CommonCredentialExtractor(ctx, pcSpec.Credentials.Source, client, pcSpec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		creds := map[string]any{}
		if err := json.Unmarshal(data, &creds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}
		cfg, err := providerConfiguration(creds)
		if err != nil {
			return ps, err
		}
		ps.Configuration = cfg
		// Served in-process (no-fork): upjet configures this provider with
		// ps.Configuration on every Connect.
		ps.FrameworkProvider = config.FrameworkProvider()
		return ps, nil
	}
}

// providerConfiguration maps the ProviderConfig credentials (a JSON object) to
// the e-breuninger/netbox provider configuration.
//
// server_url and api_token are required: the provider's Configure requires
// both unless the NETBOX_SERVER_URL / NETBOX_API_TOKEN env vars are set, which
// we do not rely on. request_timeout is optional, in seconds, as a JSON number
// or a numeric string; it defaults to defaultRequestTimeout.
func providerConfiguration(creds map[string]any) (map[string]any, error) {
	serverURL, _ := creds[keyServerURL].(string)
	apiToken, _ := creds[keyAPIToken].(string)
	if serverURL == "" || apiToken == "" {
		return nil, errors.Errorf("credentials must contain non-empty string %q and %q keys", keyServerURL, keyAPIToken)
	}

	timeout := int64(defaultRequestTimeout)
	if v, ok := creds[keyRequestTimeout]; ok {
		var err error
		if timeout, err = parseSeconds(v); err != nil {
			return nil, errors.Wrapf(err, "invalid %q in credentials", keyRequestTimeout)
		}
	}

	return map[string]any{
		keyServerURL:      serverURL,
		keyAPIToken:       apiToken,
		keyRequestTimeout: timeout,
		// upjet configures the in-process provider on every reconcile, and the
		// provider would then query NetBox's version each time; the check only
		// ever warns.
		keySkipVersion: true,
	}, nil
}

// parseSeconds accepts a positive whole number of seconds as a JSON number
// (decoded as float64) or a numeric string.
func parseSeconds(v any) (int64, error) {
	var n int64
	switch t := v.(type) {
	case float64:
		if t != float64(int64(t)) {
			return 0, errors.Errorf("not a whole number of seconds: %v", t)
		}
		n = int64(t)
	case string:
		var err error
		if n, err = strconv.ParseInt(t, 10, 64); err != nil {
			return 0, errors.Errorf("not a whole number of seconds: %q", t)
		}
	default:
		return 0, errors.Errorf("must be a number of seconds, got %T", v)
	}
	if n <= 0 {
		return 0, errors.Errorf("must be positive, got %d", n)
	}
	return n, nil
}

func toSharedPCSpec(pc *clusterv1beta1.ProviderConfig) (*namespacedv1beta1.ProviderConfigSpec, error) {
	if pc == nil {
		return nil, nil
	}
	data, err := json.Marshal(pc.Spec)
	if err != nil {
		return nil, err
	}

	var mSpec namespacedv1beta1.ProviderConfigSpec
	err = json.Unmarshal(data, &mSpec)
	return &mSpec, err
}

func resolveProviderConfig(ctx context.Context, crClient client.Client, mg resource.Managed) (*namespacedv1beta1.ProviderConfigSpec, error) {
	switch managed := mg.(type) {
	case resource.LegacyManaged: //nolint:staticcheck // still handling cluster-scoped behavior
		return resolveLegacy(ctx, crClient, managed)
	case resource.ModernManaged:
		return resolveModern(ctx, crClient, managed)
	default:
		return nil, errors.New("resource is not a managed resource")
	}
}

func resolveLegacy(ctx context.Context, client client.Client, mg resource.LegacyManaged) (*namespacedv1beta1.ProviderConfigSpec, error) { //nolint:staticcheck // still handling cluster-scoped behavior
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}
	pc := &clusterv1beta1.ProviderConfig{}
	if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	t := resource.NewLegacyProviderConfigUsageTracker(client, &clusterv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	return toSharedPCSpec(pc)
}

func resolveModern(ctx context.Context, crClient client.Client, mg resource.ModernManaged) (*namespacedv1beta1.ProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}

	pcRuntimeObj, err := crClient.Scheme().New(namespacedv1beta1.SchemeGroupVersion.WithKind(configRef.Kind))
	if err != nil {
		return nil, errors.Wrap(err, "unknown GVK for ProviderConfig")
	}
	pcObj, ok := pcRuntimeObj.(client.Object)
	if !ok {
		// This indicates a programming error, types are not properly generated
		return nil, errors.New(" is not an Object")
	}

	// Namespace will be ignored if the PC is a cluster-scoped type
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name, Namespace: mg.GetNamespace()}, pcObj); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	var pcSpec namespacedv1beta1.ProviderConfigSpec
	pcu := &namespacedv1beta1.ProviderConfigUsage{}
	switch pc := pcObj.(type) {
	case *namespacedv1beta1.ProviderConfig:
		pcSpec = pc.Spec
		if pcSpec.Credentials.SecretRef != nil {
			pcSpec.Credentials.SecretRef.Namespace = mg.GetNamespace()
		}
	case *namespacedv1beta1.ClusterProviderConfig:
		pcSpec = pc.Spec
	default:
		return nil, errors.New("unknown provider config type")
	}
	t := resource.NewProviderConfigUsageTracker(crClient, pcu)
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}
	return &pcSpec, nil
}
