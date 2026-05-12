---
page_title: "sws_notification Resource - terraform-provider-sws"
description: |-
  A managed notification channel.
---

# sws_notification

A managed notification channel.

## Example

```hcl
resource "sws_notification" "ops" { name = "ops-pager" }
```

## Argument Reference

### Required

- `name` (string) — Channel name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Channel UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_notification.<local_name> <id>
```
