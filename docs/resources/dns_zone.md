---
page_title: "sws_dns_zone Resource - terraform-provider-sws"
description: |-
  A public DNS zone (Designate).
---

# sws_dns_zone

A public DNS zone (Designate).

## Example

```hcl
resource "sws_dns_zone" "example" {
  name  = "example.com."
  email = "admin@example.com"
  ttl   = 3600
}
```

## Argument Reference

### Required

- `name` (string) — Domain name ending in a dot, e.g. `example.com.`
- `email` (string) — Zone admin email.

### Optional

- `ttl` (int) — SOA TTL in seconds. Default 3600.
- `description` (string) — Free-form description.
- `type` (string) — `PRIMARY` (default) or `SECONDARY`.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Zone UUID.
- `status` (string) — `ACTIVE` / `PENDING` / etc.


## Import

```
terraform import sws_dns_zone.<local_name> <id>
```
