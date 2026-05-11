terraform {
  required_providers {
    sws = {
      source  = "savannaacloud/sws"
      version = "~> 0.1"
    }
  }
}

provider "sws" {}

data "sws_image" "ubuntu" { name = "Ubuntu 22.04" }
data "sws_plan"  "small"  { name = "m1.small" }

resource "sws_keypair" "admin" {
  name       = "admin"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "sws_network" "app" { name = "app-net" }

resource "sws_security_group" "web" {
  name        = "web-sg"
  description = "allow 80/443 from anywhere"
}

resource "sws_instance" "web" {
  name       = "web-01"
  plan       = data.sws_plan.small.name
  image      = data.sws_image.ubuntu.id
  network_id = sws_network.app.id
  keypair    = sws_keypair.admin.name
  public_ip  = true
}

output "instance_ip" { value = sws_instance.web.ip_address }
