---
page_title: "sws_security_group_rule Resource - terraform-provider-sws"
description: |-
  A single ingress/egress rule on a security group. Rules are immutable; updates trigger replace.
---

# sws_security_group_rule

A single ingress/egress rule on a security group. Rules are immutable; updates trigger replace.

## Example

```hcl
resource "sws_security_group_rule" "https_in" {
  security_group_id = sws_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
}
```

## Argument Reference

### Required

- `security_group_id` (string) — Security group UUID.
- `direction` (string) — `ingress` or `egress`.

### Optional

- `protocol` (string) — `tcp`, `udp`, `icmp`, or null for any.
- `ethertype` (string) — `IPv4` (default) or `IPv6`.
- `port_range_min` (int) — Lower bound of the port range.
- `port_range_max` (int) — Upper bound of the port range.
- `remote_ip_prefix` (string) — Source/destination CIDR, e.g. `0.0.0.0/0`.
- `description` (string) — Free-form description.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Rule UUID.


## Import

```
terraform import sws_security_group_rule.<local_name> <id>
```
