terraform {
  required_version = ">= 1.6.0"
}

# This OpenTofu project defines mock cloud infrastructure for testing mash's
# cloud instance discovery. It uses terraform_data resources to represent
# VMs across AWS, GCP, and Azure. The state file produced by "tofu apply"
# can be processed by mash to populate the connection list.
#
# To regenerate the test state file:
#   tofu init && tofu apply -auto-approve
#   tofu show -json > ../tofu_state.json

resource "terraform_data" "aws_prod_web" {
  input = {
    resource_type = "aws_instance"
    public_ip     = "54.203.12.87"
    private_ip    = "10.0.1.25"
    name          = "ec2-prod-web-us-east"
  }
}

resource "terraform_data" "aws_staging_worker" {
  input = {
    resource_type = "aws_instance"
    public_ip     = "18.237.94.12"
    private_ip    = "10.0.1.51"
    name          = "ec2-staging-worker"
  }
}

resource "terraform_data" "gcp_data_processor" {
  input = {
    resource_type = "google_compute_instance"
    nat_ip        = "34.75.128.44"
    network_ip    = "10.128.0.8"
    name          = "gcp-data-processor"
  }
}

resource "terraform_data" "gcp_ml_training" {
  input = {
    resource_type = "google_compute_instance"
    nat_ip        = "35.229.87.15"
    network_ip    = "10.128.0.42"
    name          = "gcp-ml-training"
  }
}

resource "terraform_data" "azure_sql_server" {
  input = {
    resource_type = "azurerm_linux_virtual_machine"
    public_ip     = "20.121.44.218"
    private_ip    = "10.1.0.14"
    name          = "az-sql-server-01"
  }
}

resource "terraform_data" "azure_k8s_node" {
  input = {
    resource_type = "azurerm_linux_virtual_machine"
    public_ip     = "20.119.77.36"
    private_ip    = "10.1.0.87"
    name          = "az-k8s-node-pool-1"
  }
}
