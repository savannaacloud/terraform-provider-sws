---
page_title: "sws_floating_ip Resource - terraform-provider-sws"
description: |-
  A public (floating) IP, optionally associated with an instance.
---

# sws_floating_ip

A public (floating) IP, optionally associated with an instance.

## Example

```hcl
resource "sws_floating_ip" "web" {
  instance_id = sws_instance.web.id
}
```

## Argument Reference

### Required

_None._

### Optional

- `instance_id` (string) — Instance to associate at create time.
- `floating_network_id` (string) — External network UUID. Auto-discovered if omitted.
- `description` (string) — Free-form description.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Floating IP UUID.
- `address` (string) — Allocated IPv4 address.


## Import

```
terraform import sws_floating_ip.<local_name> <id>
```
