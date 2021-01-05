
variable "user_data" {
  description = "cloudinit user data"
}

variable "server_name" {
  description = "human readable name of server"
}

variable "instance_size" {
  description = "Digital ocean droplet instance size. e.g. s-2vcpu-4gb (see https://slugs.do-api.dev/ for whole list)"
}

variable "region" {
  description = "Region to launch the droplet. e.g. nyc1 (see https://slugs.do-api.dev/ for whole list)"
}
