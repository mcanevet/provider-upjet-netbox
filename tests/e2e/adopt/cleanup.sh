#!/bin/sh
# Removes what adopt.sh created, whether or not it passed: the MRs (their
# deletion policy leaves NetBox alone), then the NetBox objects. Deleting the
# device deletes its interfaces and their MAC addresses.
set -u
k=${KUBECTL:-kubectl}
api=../netbox-api.sh
state=${TMPDIR:-/tmp}/e2e-adopt.env

g=dcim.netbox.m.crossplane.io
$k -n "$NAMESPACE" delete --ignore-not-found --wait --timeout=2m \
  "deviceinterfaceprimarymacaddresses.$g/e2e-adopt-eno2" \
  "macaddresses.$g/e2e-adopt-eno1" "macaddresses.$g/e2e-adopt-eno2" \
  "deviceinterfaces.$g/e2e-adopt-eno1" "deviceinterfaces.$g/e2e-adopt-eno2" \
  "devices.$g/e2e-adopt" >/dev/null 2>&1 || true

[ -f "$state" ] || exit 0
get() { sed -n "s/^$1=//p" "$state"; }
for obj in "dcim/devices/$(get device)" "dcim/device-roles/$(get role)" "dcim/device-types/$(get dtype)" "dcim/manufacturers/$(get mfr)" "dcim/sites/$(get site)"; do
  case "$obj" in */) continue ;; esac
  $api DELETE "$obj/" >/dev/null 2>&1 || echo "could not delete $obj"
done
rm -f "$state"
