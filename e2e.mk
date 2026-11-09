# ====================================================================================
# NetBox end-to-end testing
#
# Additions to the upjet-provider-template's e2e targets, kept out of the
# Makefile so that it stays close to the template. Included at the end of the
# Makefile's "End to End Testing" section, after k8s_tools.mk (for KUBECTL and
# CHAINSAW) and the template's local-deploy, uptest and e2e targets.
#
#   make e2e = local-deploy -> netbox.deploy -> uptest -> e2e-scenarios
#
# netbox.deploy is a prerequisite of uptest rather than of e2e: make builds a
# target's prerequisites in the order its rules are read, so appending it to the
# template's "e2e: local-deploy uptest" would run it after uptest.

# NetBox version the e2e tests run against. The Terraform provider v6.0.0
# supports "v4.6.8 - 4.6.10" (its README.md), which is the default here. CI also
# runs v4.4.10, outside that range: the provider only warns on an unsupported
# version (internal/provider/provider_gen.go), so whether it works is what the
# tests find out.
NETBOX_VERSION ?= v4.6.10

# Throwaway API token of the in-cluster NetBox (cluster/test/netbox.yaml).
# NetBox 4.5+ creates a v2 token (nbt_<key>.<token>, sent as "Bearer"); 4.4
# only has v1 tokens, so it uses the bare 40-character token ("Token").
NETBOX_TEST_TOKEN := Iw2JZtpMdYjBuIVIZLItd4lB2RYEiv6xhY23W0tQ
export NETBOX_TOKEN ?= $(if $(filter v4.4.% v4.3.%,$(NETBOX_VERSION)),$(NETBOX_TEST_TOKEN),nbt_mrcLJrER9ves.$(NETBOX_TEST_TOKEN))

# The ProviderConfig secret content (cluster/test/setup.sh): JSON with
# server_url and api_token. Override it to run the e2e tests against another
# NetBox.
export UPTEST_CLOUD_CREDENTIALS ?= {"server_url":"http://netbox.netbox.svc.cluster.local:8080","api_token":"$(NETBOX_TOKEN)"}

# All hand-written examples (excluding ProviderConfigs) by default.
UPTEST_EXAMPLE_LIST ?= $(shell find examples/cluster examples/namespaced -path '*/v1alpha1/*.yaml' | sort | paste -sd, -)

netbox.deploy: $(KUBECTL)
	@$(INFO) deploying NetBox $(NETBOX_VERSION) for e2e tests
	@sed 's|NETBOX_VERSION|$(NETBOX_VERSION)|' cluster/test/netbox.yaml | $(KUBECTL) apply -f -
	@$(KUBECTL) -n netbox rollout status deployment/postgres deployment/redis --timeout=5m
	@$(KUBECTL) -n netbox rollout status deployment/netbox --timeout=20m
	@$(OK) deploying NetBox $(NETBOX_VERSION) for e2e tests

# Chainsaw scenarios of tests/e2e that change NetBox behind the provider's back
# (uptest only drives the provider). Needs the provider and NetBox deployed
# (local-deploy netbox.deploy); creates the default ProviderConfigs itself.
e2e-scenarios: $(KUBECTL) $(CHAINSAW)
	@$(INFO) running e2e scenarios
	@KUBECTL=$(KUBECTL) cluster/test/setup.sh
	@KUBECTL=$(KUBECTL) $(CHAINSAW) test tests/e2e --config tests/e2e/.chainsaw.yaml || $(FAIL)
	@$(OK) running e2e scenarios

uptest: netbox.deploy
e2e: e2e-scenarios

# Load test (tests/load/run.sh): LOAD_COUNT Site MRs against the in-cluster
# NetBox, timing create, steady state and delete, and sampling the provider's
# CPU/memory, reconcile rate and NetBox API calls. Tunables (environment):
# LOAD_COUNT, LOAD_MAX_RECONCILE_RATE, LOAD_POLL, LOAD_STEADY_SECONDS.
# Results in _output/load.
load-test: $(KUBECTL)
	@$(INFO) running load test
	@KUBECTL=$(KUBECTL) cluster/test/setup.sh
	@KUBECTL=$(KUBECTL) NETBOX_VERSION=$(NETBOX_VERSION) LOAD_OUTPUT_DIR=$(OUTPUT_DIR)/load tests/load/run.sh || $(FAIL)
	@$(OK) running load test

load: local-deploy netbox.deploy load-test

.PHONY: netbox.deploy e2e-scenarios load-test load
