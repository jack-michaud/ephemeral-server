provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_volume" "mc_vol" {
  name   = "minecraft-volume-${var.server_name}"
  region = "nyc1"
  size = 20
  initial_filesystem_type = "ext4"
}


# Create a new Web Droplet in the nyc2 region
resource "digitalocean_droplet" "minecraft" {
  image  = "ubuntu-18-04-x64"
  name   = "minecraft-server-${var.server_name}"
  region = "nyc1"
  size   = "s-2vcpu-4gb"

  user_data = templatefile("${path.module}/digitalocean_user_data.tmpl", {
    public_key_data = file(var.public_key_path)
  })

}

resource "digitalocean_volume_attachment" "minecraft-vol-attach" {
  droplet_id = digitalocean_droplet.minecraft.id
  volume_id = digitalocean_volume.mc_vol.id
}
