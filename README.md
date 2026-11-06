# provider-upjet-netbox

`provider-upjet-netbox` is a [Crossplane](https://crossplane.io/) provider for
[NetBox](https://netboxlabs.com/), built with
[Upjet](https://github.com/crossplane/upjet) code generation on top of the
[e-breuninger/netbox](https://github.com/e-breuninger/terraform-provider-netbox)
Terraform provider (v6.0.0).

> **Status: experimental.** Only a small allowlist of resources is generated
> (see `config/external_name.go`). It has not been tested against a live NetBox.

## Generated resources

| Terraform resource       | API group                         | Kind             |
|--------------------------|-----------------------------------|------------------|
| `netbox_device`          | `dcim.netbox.crossplane.io`       | `Device`         |
| `netbox_device_role`     | `dcim.netbox.crossplane.io`       | `DeviceRole`     |
| `netbox_site`            | `dcim.netbox.crossplane.io`       | `Site`           |
| `netbox_tenant`          | `tenancy.netbox.crossplane.io`    | `Tenant`         |
| `netbox_ip_address`      | `ipam.netbox.crossplane.io`       | `IPAddress`      |
| `netbox_prefix`          | `ipam.netbox.crossplane.io`       | `Prefix`         |
| `netbox_vlan`            | `ipam.netbox.crossplane.io`       | `VLAN`           |
| `netbox_cluster`         | `virtualization.netbox.crossplane.io` | `Cluster`    |
| `netbox_virtual_machine` | `virtualization.netbox.crossplane.io` | `VirtualMachine` |

Namespaced variants live under `*.netbox.m.crossplane.io`.

## Credentials

The ProviderConfig secret must hold a JSON document:

```json
{"server_url": "https://netbox.example.com", "api_token": "..."}
```

## Developing

```console
make submodules   # first time only
make generate     # regenerate CRDs, types and controllers
make build
```

To add a resource, add it to `ExternalNameConfigs` in
`config/external_name.go` (and to `config/groups.go`) and run `make generate`.
