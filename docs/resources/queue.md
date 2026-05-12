---
page_title: "sws_queue Resource - terraform-provider-sws"
description: |-
  A managed message queue.
---

# sws_queue

A managed message queue.

## Example

```hcl
resource "sws_queue" "jobs" { name = "jobs" }
```

## Argument Reference

### Required

- `name` (string) — Queue name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Queue UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_queue.<local_name> <id>
```
