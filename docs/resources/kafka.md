---
page_title: "sws_kafka Resource - terraform-provider-sws"
description: |-
  A managed Kafka cluster.
---

# sws_kafka

A managed Kafka cluster.

## Example

```hcl
resource "sws_kafka" "events" { name = "events" }
```

## Argument Reference

### Required

- `name` (string) — Cluster name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Cluster UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_kafka.<local_name> <id>
```
