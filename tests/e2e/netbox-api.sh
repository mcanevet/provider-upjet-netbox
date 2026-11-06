#!/bin/sh
# Calls the REST API of the e2e NetBox (cluster/test/netbox.yaml) from inside
# its pod, so the tests can change NetBox behind the provider's back.
#
#   netbox-api.sh METHOD PATH [JSON_BODY]
#   netbox-api.sh GET    dcim/sites/12/
#   netbox-api.sh PATCH  dcim/sites/12/ '{"description":"changed"}'
#   netbox-api.sh DELETE dcim/sites/12/
#
# Exits non-zero on an HTTP error (curl -f), e.g. 404 for a missing object.
# NETBOX_TOKEN defaults to the throwaway v2 token of cluster/test/netbox.yaml.
set -eu

method=$1
path=$2
body=${3:-}
token=${NETBOX_TOKEN:-nbt_mrcLJrER9ves.Iw2JZtpMdYjBuIVIZLItd4lB2RYEiv6xhY23W0tQ}
kubectl=${KUBECTL:-kubectl}

set -- -sS -f -X "$method" \
  -H "Authorization: Bearer $token" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json"
if [ -n "$body" ]; then
  set -- "$@" --data "$body"
fi

exec "$kubectl" -n netbox exec deploy/netbox -c netbox -- \
  curl "$@" "http://localhost:8080/api/$path"
