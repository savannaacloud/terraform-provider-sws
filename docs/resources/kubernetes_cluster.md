---
page_title: "sws_kubernetes_cluster Resource - terraform-provider-sws"
description: |-
  A Magnum Kubernetes cluster. Polls CREATE_COMPLETE (~8-15 min).
---

# sws_kubernetes_cluster

A Magnum Kubernetes cluster. Polls CREATE_COMPLETE (~8-15 min).

## Example

```hcl
resource "sws_kubernetes_cluster" "prod" {
  name                = "prod"
  cluster_template_id = sws_kubernetes_template.k8s28.id
  node_count          = 3
  master_count        = 1
  keypair_id          = sws_keypair.admin.name
}
```

## Argument Reference

### Required

- `name` (string) — Cluster name.
- `cluster_template_id` (string) — Template to derive from.
- `node_count` (int) — Worker node count.

### Optional

- `master_count` (int) — Master node count. Default 1.
- `keypair_id` (string) — Override keypair.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Cluster UUID.
- `status` (string) — `CREATE_IN_PROGRESS` / `CREATE_COMPLETE` / `CREATE_FAILED`.
- `api_address` (string) — k8s API endpoint URL.


## Import

```
terraform import sws_kubernetes_cluster.<local_name> <id>
```
