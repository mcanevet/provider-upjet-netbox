#!/usr/bin/env bash
# Load test: how many managed resources can one provider create and keep
# reconciled, and at what CPU/memory and NetBox API cost?
#
# It applies LOAD_COUNT Site MRs (a resource without references) and times
# three phases:
#   create  - until all are Ready
#   steady  - LOAD_STEADY_SECONDS with everything Ready, drift-checked every
#             LOAD_POLL: does the provider keep up, and at what cost?
#   delete  - until all are gone
# Every 15-30s it samples Ready/Synced counts, provider and NetBox CPU/memory
# (metrics-server), and the provider's reconcile count, workqueue depth and
# external API calls.
#
# Expects `make local-deploy netbox.deploy` (a provider and the in-cluster
# NetBox), KUBECTL and jq. Writes samples.csv and summary.md to
# LOAD_OUTPUT_DIR.
set -euo pipefail

: "${LOAD_COUNT:=1000}"
: "${LOAD_MAX_RECONCILE_RATE:=10}"
: "${LOAD_POLL:=10m}"
: "${LOAD_STEADY_SECONDS:=600}"
: "${LOAD_TIMEOUT_SECONDS:=7200}"
: "${LOAD_OUTPUT_DIR:=_output/load}"
: "${METRICS_SERVER_VERSION:=v0.9.1}"
: "${KUBECTL:=kubectl}"
k=$KUBECTL
ns=load
provider_ns=crossplane-system
drc=runtimeconfig-provider-upjet-netbox
mrs=sites.dcim.netbox.m.crossplane.io

mkdir -p "$LOAD_OUTPUT_DIR"
samples="$LOAD_OUTPUT_DIR/samples.csv"
summary="$LOAD_OUTPUT_DIR/summary.md"
echo "phase,elapsed_s,mrs,ready,synced_false,provider_cpu_m,provider_mem_mi,netbox_cpu_m,reconciles_total,workqueue_depth,ext_api_calls" > "$samples"

log() { echo "$(date -u +%H:%M:%S) $*"; }

# --- environment ------------------------------------------------------------

log "installing metrics-server $METRICS_SERVER_VERSION"
$k apply -f "https://github.com/kubernetes-sigs/metrics-server/releases/download/$METRICS_SERVER_VERSION/components.yaml" >/dev/null
# kind kubelets serve self-signed certificates.
$k -n kube-system patch deployment metrics-server --type=json \
  -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]' >/dev/null
$k -n kube-system rollout status deployment/metrics-server --timeout=5m >/dev/null

log "configuring the provider: --max-reconcile-rate=$LOAD_MAX_RECONCILE_RATE --poll=$LOAD_POLL (no --debug)"
$k patch deploymentruntimeconfig "$drc" --type=json -p "[{\"op\":\"replace\",\"path\":\"/spec/deploymentTemplate/spec/template/spec/containers/0/args\",\"value\":[\"--max-reconcile-rate=$LOAD_MAX_RECONCILE_RATE\",\"--poll=$LOAD_POLL\"]}]" >/dev/null
sleep 10
$k -n "$provider_ns" rollout status deployment -l pkg.crossplane.io/provider=provider-upjet-netbox --timeout=5m >/dev/null
provider_pod() { $k -n "$provider_ns" get pod -l pkg.crossplane.io/provider=provider-upjet-netbox -o jsonpath='{.items[0].metadata.name}'; }

$k create namespace "$ns" --dry-run=client -o yaml | $k apply -f - >/dev/null

# --- managed resources ----------------------------------------------------------

manifest="$LOAD_OUTPUT_DIR/mrs.yaml"
for ((i = 0; i < LOAD_COUNT; i++)); do
  cat <<EOF
---
apiVersion: dcim.netbox.m.crossplane.io/v1alpha1
kind: Site
metadata:
  name: load-$i
  namespace: $ns
spec:
  forProvider:
    name: load-$i
    slug: load-$i
    status: active
  providerConfigRef:
    kind: ClusterProviderConfig
    name: default
EOF
done > "$manifest"

# --- sampling -------------------------------------------------------------------

t0=0
sample() { # phase
  local now elapsed json count ready unsynced pod metrics ptop ntop
  now=$(date +%s); elapsed=$((now - t0))
  json=$($k -n "$ns" get "$mrs" -o json 2>/dev/null || echo '{"items":[]}')
  count=$(jq '.items | length' <<<"$json")
  ready=$(jq '[.items[] | select(any(.status.conditions[]?; .type == "Ready" and .status == "True"))] | length' <<<"$json")
  unsynced=$(jq '[.items[] | select(any(.status.conditions[]?; .type == "Synced" and .status == "False"))] | length' <<<"$json")
  pod=$(provider_pod)
  metrics=$($k get --raw "/api/v1/namespaces/$provider_ns/pods/$pod:8080/proxy/metrics" 2>/dev/null || true)
  # m METRIC [SUBSTRING]: sum of the samples of METRIC whose labels contain
  # SUBSTRING (label names differ between controller-runtime versions).
  m() { awk -v name="$1" -v sub_="${2:-}" '(index($0, name "{") == 1 || index($0, name " ") == 1) && index($0, sub_) { s += $NF } END { printf "%d", s }' <<<"$metrics"; }
  ptop=$($k -n "$provider_ns" top pod "$pod" --no-headers 2>/dev/null | awk '{print $2","$3}' | tr -d 'mMi' || true)
  ntop=$($k -n netbox top pod -l app=netbox --no-headers 2>/dev/null | awk '{print $2}' | tr -d 'm' || true)
  echo "$1,$elapsed,$count,$ready,$unsynced,${ptop:-,},${ntop:-},$(m controller_runtime_reconcile_total 'kind=site'),$(m workqueue_depth 'kind=site'),$(m upjet_resource_ext_api_duration_count)" | tee -a "$samples"
}
wait_until() { # phase, condition (evaluated after each sample), timeout
  local deadline=$(($(date +%s) + $3))
  while :; do
    line=$(sample "$1")
    eval "$2" && return 0
    [ "$(date +%s)" -lt "$deadline" ] || { log "$1: timed out"; return 1; }
    sleep 15
  done
}
field() { cut -d, -f"$1" <<<"$line"; }

# --- phases ---------------------------------------------------------------------

log "create: applying $LOAD_COUNT Sites"
t0=$(date +%s)
$k apply -f "$manifest" >/dev/null
wait_until create '[ "$(field 3)" -eq '"$LOAD_COUNT"' ] && [ "$(field 4)" -eq '"$LOAD_COUNT"' ]' "$LOAD_TIMEOUT_SECONDS"
create_s=$(($(date +%s) - t0))
log "create: $LOAD_COUNT Ready in ${create_s}s"

log "steady: ${LOAD_STEADY_SECONDS}s with everything Ready (poll $LOAD_POLL)"
line=$(sample steady); r0=$(field 9); a0=$(field 11); s0=$(date +%s)
end=$(($(date +%s) + LOAD_STEADY_SECONDS))
while [ "$(date +%s)" -lt "$end" ]; do sleep 30; line=$(sample steady); done
elapsed=$(($(date +%s) - s0))
steady_rate=$(awk -v a="$r0" -v b="$(field 9)" -v t="$elapsed" 'BEGIN { printf "%.2f", (b - a) / t }')
api_rate=$(awk -v a="$a0" -v b="$(field 11)" -v t="$elapsed" 'BEGIN { printf "%.2f", (b - a) / t }')
not_ready_after=$((LOAD_COUNT - $(field 4)))

log "delete: deleting $LOAD_COUNT Sites"
t0=$(date +%s)
$k -n "$ns" delete "$mrs" --all --wait=false >/dev/null
wait_until delete '[ "$(field 3)" -eq 0 ]' "$LOAD_TIMEOUT_SECONDS"
delete_s=$(($(date +%s) - t0))

# --- summary --------------------------------------------------------------------

peak() { awk -F, -v p="$1" -v c="$2" '$1 == p && $c != "" && $c + 0 > m { m = $c + 0 } END { print m + 0 }' "$samples"; }
needed=$(awk -v n="$LOAD_COUNT" -v p="$LOAD_POLL" 'BEGIN { s = p + 0; if (p ~ /m$/) s *= 60; if (p ~ /h$/) s *= 3600; printf "%.2f", n / s }')
cat > "$summary" <<EOF
## Load test: $LOAD_COUNT Sites

Provider: --max-reconcile-rate=$LOAD_MAX_RECONCILE_RATE --poll=$LOAD_POLL; NetBox ${NETBOX_VERSION:-?}.

| Phase | Result |
|---|---|
| create | all $LOAD_COUNT Ready in **${create_s}s** ($(awk -v n="$LOAD_COUNT" -v t="$create_s" 'BEGIN { printf "%.1f", n / t }') MR/s) |
| steady (${LOAD_STEADY_SECONDS}s) | **$steady_rate reconciles/s** (needed: $needed/s), **$api_rate external API calls/s**; $not_ready_after not Ready at the end |
| delete | all gone in **${delete_s}s** |

| Peak | create | steady | delete |
|---|---|---|---|
| provider CPU (millicores) | $(peak create 6) | $(peak steady 6) | $(peak delete 6) |
| provider memory (MiB) | $(peak create 7) | $(peak steady 7) | $(peak delete 7) |
| NetBox CPU (millicores) | $(peak create 8) | $(peak steady 8) | $(peak delete 8) |
| site workqueue depth | $(peak create 10) | $(peak steady 10) | $(peak delete 10) |

Samples every 15-30s: samples.csv.
EOF
cat "$summary"
