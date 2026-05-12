---
page_title: "sws_private_dns_zone Resource - terraform-provider-sws"
description: |-
  A private DNS zone (resolved only inside your project networks).
---

# sws_private_dns_zone

A private DNS zone (resolved only inside your project networks).

## Example

```hcl
resource "sws_private_dns_zone" "internal" {
  name        = "internal.local"
  description = "service discovery"
}
```

## Argument Reference

### Required

- `name` (string) — Zone name (no trailing dot needed).

### Optional

- `description` (string) — Free-form description.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Zone UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_private_dns_zone.<local_name> <id>
```
