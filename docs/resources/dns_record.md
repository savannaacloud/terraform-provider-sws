---
page_title: "sws_dns_record Resource - terraform-provider-sws"
description: |-
  A recordset in a public DNS zone.
---

# sws_dns_record

A recordset in a public DNS zone.

## Example

```hcl
resource "sws_dns_record" "www" {
  zone_id = sws_dns_zone.example.id
  name    = "www.example.com."
  type    = "A"
  ttl     = 300
  records = ["1.2.3.4"]
}
```

## Argument Reference

### Required

- `zone_id` (string) — Parent zone UUID.
- `name` (string) — Fully qualified name ending in a dot.
- `type` (string) — `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`.
- `records` (list(string)) — Record values, e.g. `["1.2.3.4"]`.

### Optional

- `ttl` (int) — Record TTL.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Composite `<zone_id>:<rrset_id>`.


## Import

```
terraform import sws_dns_record.<local_name> <id>
```

For composite-ID resources, the import id is shown in the Read-only `id` attribute (e.g. `<router_id>:<subnet_id>`).
