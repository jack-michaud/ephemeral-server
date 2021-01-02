output "ip" {
  value = digitalocean_droplet.minecraft.ipv4_address
}
output "public_key" {
  value = var.public_key
}
