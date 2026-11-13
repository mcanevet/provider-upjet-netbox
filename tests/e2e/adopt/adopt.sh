#!/usr/bin/env bash
# See chainsaw-test.yaml. Creates in NetBox what Purchase, HW&DC and the old
# inventory sync would have set, applies the MRs a composition would render
# for one device, and checks that NetBox only gained what the MRs create.
set -eu
k=${KUBECTL:-kubectl}
ns=$NAMESPACE
api=../netbox-api.sh
state=${TMPDIR:-/tmp}/e2e-adopt.env
: > "$state"

# id METHOD PATH JSON: the id of the object created by the request.
id() { $api "$@" | jq -er .id; }
remember() { echo "$1=$2" >> "$state"; }
fail() { echo "FAIL: $*"; exit 1; }

# --- what exists before Crossplane ----------------------------------------------

site=$(id POST dcim/sites/ '{"name":"e2e-adopt","slug":"e2e-adopt","status":"active"}'); remember site "$site"
mfr=$(id POST dcim/manufacturers/ '{"name":"e2e-adopt","slug":"e2e-adopt"}'); remember mfr "$mfr"
dtype=$(id POST dcim/device-types/ "{\"manufacturer\":$mfr,\"model\":\"e2e-adopt\",\"slug\":\"e2e-adopt\"}"); remember dtype "$dtype"
role=$(id POST dcim/device-roles/ '{"name":"e2e-adopt","slug":"e2e-adopt","color":"9e9e9e"}'); remember role "$role"
# Purchase and HW&DC.
device=$(id POST dcim/devices/ "{\"name\":\"e2e-adopt\",\"device_type\":$dtype,\"role\":$role,\"site\":$site,\"status\":\"planned\",\"serial\":\"HWDC-SERIAL\",\"asset_tag\":\"e2e-adopt-tag\",\"comments\":\"set by HW&DC\"}"); remember device "$device"
bmc=$(id POST dcim/interfaces/ "{\"device\":$device,\"name\":\"bmc0\",\"type\":\"1000base-t\",\"mgmt_only\":true}")
# The old sync: eno1 with its network settings and MAC address.
eno1=$(id POST dcim/interfaces/ "{\"device\":$device,\"name\":\"eno1\",\"type\":\"1000base-t\",\"mtu\":9000,\"mode\":\"access\",\"description\":\"set by the old sync\"}")
eno1_mac=$(id POST dcim/mac-addresses/ "{\"mac_address\":\"02:00:00:00:AA:01\",\"assigned_object_type\":\"dcim.interface\",\"assigned_object_id\":$eno1}")
$api PATCH "dcim/interfaces/$eno1/" "{\"primary_mac_address\":$eno1_mac}" >/dev/null

# The fields others set, plus last_updated, which changes on any write, even
# one that sends the same values. Counters (e.g. a device's interface_count)
# are left out: they change when eno2 is created.
before() {
  case "$1" in
    dcim/devices/*) f='{name, status: .status.value, serial, asset_tag, comments, tenant, platform, primary_ip4, oob_ip, custom_fields, last_updated}' ;;
    dcim/interfaces/*) f='{name, type: .type.value, label, mtu, mode, description, mgmt_only, enabled, primary_mac_address: .primary_mac_address.id, last_updated}' ;;
    dcim/mac-addresses/*) f='{mac_address, assigned_object_type, assigned_object_id, description, last_updated}' ;;
  esac
  $api GET "$1" | jq -S "$f"
}
device_before=$(before "dcim/devices/$device/")
bmc_before=$(before "dcim/interfaces/$bmc/")
eno1_before=$(before "dcim/interfaces/$eno1/")
eno1_mac_before=$(before "dcim/mac-addresses/$eno1_mac/")
echo "NetBox: device $device, bmc0 $bmc, eno1 $eno1 (MAC $eno1_mac)"

# --- what a composition renders for the device ----------------------------------

$k -n "$ns" apply -f - <<EOF
apiVersion: dcim.netbox.m.crossplane.io/v1alpha1
kind: Device
metadata:
  name: e2e-adopt
  annotations:
    crossplane.io/external-name: "$device"
spec:
  managementPolicies: ["Observe"]
  forProvider: {}
  providerConfigRef: {kind: ClusterProviderConfig, name: default}
---
# Exists: adopted by ID. Its label differs from NetBox's on purpose: without
# the Update policy it must not be written.
apiVersion: dcim.netbox.m.crossplane.io/v1alpha1
kind: DeviceInterface
metadata:
  name: e2e-adopt-eno1
  annotations:
    crossplane.io/external-name: "$eno1"
spec:
  managementPolicies: ["Observe", "Create"]
  deletionPolicy: Orphan
  forProvider:
    name: eno1
    type: 1000base-t
    label: NIC.Integrated.1-1
    deviceIdRef: {name: e2e-adopt}
  providerConfigRef: {kind: ClusterProviderConfig, name: default}
---
apiVersion: dcim.netbox.m.crossplane.io/v1alpha1
kind: MACAddress
metadata:
  name: e2e-adopt-eno1
  annotations:
    crossplane.io/external-name: "$eno1_mac"
spec:
  managementPolicies: ["Observe", "Create"]
  deletionPolicy: Orphan
  forProvider:
    macAddress: "02:00:00:00:AA:01"
    deviceInterfaceIdRef: {name: e2e-adopt-eno1}
  providerConfigRef: {kind: ClusterProviderConfig, name: default}
---
# Missing: created.
apiVersion: dcim.netbox.m.crossplane.io/v1alpha1
kind: DeviceInterface
metadata:
  name: e2e-adopt-eno2
spec:
  managementPolicies: ["Observe", "Create"]
  deletionPolicy: Orphan
  forProvider:
    name: eno2
    type: 1000base-t
    label: NIC.Integrated.1-2
    deviceIdRef: {name: e2e-adopt}
  providerConfigRef: {kind: ClusterProviderConfig, name: default}
---
apiVersion: dcim.netbox.m.crossplane.io/v1alpha1
kind: MACAddress
metadata:
  name: e2e-adopt-eno2
spec:
  managementPolicies: ["Observe", "Create"]
  deletionPolicy: Orphan
  forProvider:
    macAddress: "02:00:00:00:AA:02"
    deviceInterfaceIdRef: {name: e2e-adopt-eno2}
  providerConfigRef: {kind: ClusterProviderConfig, name: default}
---
apiVersion: dcim.netbox.m.crossplane.io/v1alpha1
kind: DeviceInterfacePrimaryMACAddress
metadata:
  name: e2e-adopt-eno2
spec:
  deletionPolicy: Orphan
  forProvider:
    deviceInterfaceIdRef: {name: e2e-adopt-eno2}
    macAddressIdRef: {name: e2e-adopt-eno2}
  providerConfigRef: {kind: ClusterProviderConfig, name: default}
EOF

mrs="devices.dcim.netbox.m.crossplane.io/e2e-adopt
deviceinterfaces.dcim.netbox.m.crossplane.io/e2e-adopt-eno1
macaddresses.dcim.netbox.m.crossplane.io/e2e-adopt-eno1
deviceinterfaces.dcim.netbox.m.crossplane.io/e2e-adopt-eno2
macaddresses.dcim.netbox.m.crossplane.io/e2e-adopt-eno2
deviceinterfaceprimarymacaddresses.dcim.netbox.m.crossplane.io/e2e-adopt-eno2"
for mr in $mrs; do
  if ! $k -n "$ns" wait --for=condition=Ready --timeout=4m "$mr" >/dev/null; then
    $k -n "$ns" get "$mr" -o jsonpath='{range .status.conditions[*]}{.type}={.status} {.reason}: {.message}{"\n"}{end}'
    fail "$mr is not Ready"
  fi
  $k -n "$ns" wait --for=condition=Synced --timeout=1m "$mr" >/dev/null || fail "$mr is not Synced"
done
echo "all MRs Ready and Synced"

en() { $k -n "$ns" get "$1" -o jsonpath='{.metadata.annotations.crossplane\.io/external-name}'; }

# --- adopted: same objects, unchanged ---------------------------------------------

[ "$(en deviceinterfaces.dcim.netbox.m.crossplane.io/e2e-adopt-eno1)" = "$eno1" ] || fail "eno1 was not adopted"
[ "$(en macaddresses.dcim.netbox.m.crossplane.io/e2e-adopt-eno1)" = "$eno1_mac" ] || fail "eno1's MAC address was not adopted"
# Give a Create-only MR a few reconciles to (wrongly) write anything.
for i in 1 2 3; do
  for mr in $mrs; do $k -n "$ns" annotate --overwrite "$mr" e2e.netbox.crossplane.io/nudge="$i" >/dev/null; done
  sleep 10
done
same() { # what path before
  now=$(before "$2")
  if [ "$now" != "$3" ]; then
    echo "$1 changed in NetBox:"; diff <(echo "$3") <(echo "$now") || true
    fail "$1 changed"
  fi
}
same device "dcim/devices/$device/" "$device_before"
same bmc0 "dcim/interfaces/$bmc/" "$bmc_before"
same eno1 "dcim/interfaces/$eno1/" "$eno1_before"
same "eno1's MAC address" "dcim/mac-addresses/$eno1_mac/" "$eno1_mac_before"
echo "device, bmc0, eno1 and its MAC address unchanged"

# --- created: what the MRs set ----------------------------------------------------

eno2=$(en deviceinterfaces.dcim.netbox.m.crossplane.io/e2e-adopt-eno2)
eno2_mac=$(en macaddresses.dcim.netbox.m.crossplane.io/e2e-adopt-eno2)
case "$eno2$eno2_mac" in ''|*[!0-9]*) fail "eno2 or its MAC address has no NetBox ID: '$eno2' '$eno2_mac'" ;; esac
eno2_json=$($api GET "dcim/interfaces/$eno2/")
[ "$(echo "$eno2_json" | jq -r '.device.id, .name, .label' | tr '\n' ' ')" = "$device eno2 NIC.Integrated.1-2 " ] || fail "eno2 is not as created: $(echo "$eno2_json" | jq -c '{device: .device.id, name, label}')"
[ "$(echo "$eno2_json" | jq -r '.primary_mac_address.id')" = "$eno2_mac" ] || fail "eno2's primary MAC address is not $eno2_mac"
[ "$($api GET "dcim/mac-addresses/$eno2_mac/" | jq -r '.assigned_object_id')" = "$eno2" ] || fail "MAC address $eno2_mac is not assigned to eno2"
[ "$($api GET "dcim/interfaces/?device_id=$device&limit=0" | jq -r .count)" = 3 ] || fail "the device does not have exactly bmc0, eno1 and eno2"
echo "eno2 $eno2 created with MAC address $eno2_mac as its primary one"

# --- deleting the MRs leaves NetBox as it is -------------------------------------

for mr in $mrs; do $k -n "$ns" delete --wait --timeout=3m "$mr" >/dev/null; done
for path in "dcim/devices/$device/" "dcim/interfaces/$eno1/" "dcim/interfaces/$eno2/" "dcim/mac-addresses/$eno1_mac/" "dcim/mac-addresses/$eno2_mac/"; do
  $api GET "$path" >/dev/null || fail "$path is gone after deleting the MRs"
done
[ "$($api GET "dcim/interfaces/$eno2/" | jq -r '.primary_mac_address.id')" = "$eno2_mac" ] || fail "eno2's primary MAC address was cleared"
echo "MRs deleted, NetBox unchanged"
