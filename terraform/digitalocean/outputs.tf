output "ip" {
  value = digitalocean_droplet.minecraft.ipv4_address
}
output "permanent_device" {
  value = "/dev/sda"
}
