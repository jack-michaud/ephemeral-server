output "ip" {
  value = var.cloud_provider == "digitalocean" ? module.digitalocean[0].ip : module.aws[0].ip
}
output "permanent_device" {
  value = var.cloud_provider == "digitalocean" ? "/dev/sda" : "/dev/nvme1n1"
}
