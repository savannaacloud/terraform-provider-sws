---
page_title: "sws_subnet Resource - terraform-provider-sws"
description: |-
  An IPv4 subnet on an existing network.
---

# sws_subnet

An IPv4 subnet on an existing network.

## Example

```hcl
resource "sws_subnet" "extra" {
  network_id = sws_network.app.id
  name       = "extra-subnet"
  cidr       = "10.0.1.0/24"
}
```

## Argument Reference

### Required

- `network_id` (string) — Network UUID.
- `name` (string) — Display name.
- `cidr` (string) — IPv4 CIDR.

### Optional

- `ip_version` (int) — Defaults to 4.
- `gateway_ip` (string) — Defaults to the first usable IP.
- `enable_dhcp` (bool) — Defaults to true.
- `dns_nameservers` (list(string)) — Defaults to `["1.1.1.1", "8.8.8.8"]`.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Subnet UUID.


## Import

```
terraform import sws_subnet.<local_name> <id>
```
