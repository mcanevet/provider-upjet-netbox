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
# NETBOX_TOKEN defaults to the throwaway v2 token of cluster/test/netbox.yaml;
# `make` exports the one matching NETBOX_VERSION. v2 tokens (nbt_ prefix) are
# sent as "Bearer", v1 tokens (NetBox 4.4 and older) as "Token", like the
# Terraform provider does.
set -eu

method=$1
path=$2
body=${3:-}
token=${NETBOX_TOKEN:-nbt_mrcLJrER9ves.Iw2JZtpMdYjBuIVIZLItd4lB2RYEiv6xhY23W0tQ}
kubectl=${KUBECTL:-kubectl}
case "$token" in
  nbt_*) scheme=Bearer ;;
  *) scheme=Token ;;
esac

set -- -sS -f -X "$method" \
  -H "Authorization: $scheme $token" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json"
if [ -n "$body" ]; then
  set -- "$@" --data "$body"
fi

exec "$kubectl" -n netbox exec deploy/netbox -c netbox -- \
  curl "$@" "http://localhost:8080/api/$path"
