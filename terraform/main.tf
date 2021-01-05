provider "digitalocean" {
  token = var.do_token
}

provider "aws" {
  region = "us-east-1"
}

locals {
  user_data = templatefile("${path.module}/user_data.tmpl", {
    public_key_data = file(var.public_key_path)
  })
}

module "digitalocean" {
  count = var.cloud_provider == "digitalocean" ? 1 : 0
  source = "./digitalocean"
  server_name = var.server_name
  user_data = local.user_data
  
  instance_size = "s-2vcpu-4gb" 
  region = "nyc1"
}

module "aws" {
  count = var.cloud_provider == "aws" ? 1 : 0
  source = "./aws"
  server_name = var.server_name
  user_data = local.user_data
  
  instance_size = "t3.large" 
  region = "us-east-1a"
}
