variable "server_name" {
  description = "human readable name of server"
}

variable "instance_size" {
  description = "Digital ocean droplet instance size. e.g. s-2vcpu-4gb (see https://slugs.do-api.dev/ for whole list)"
}

variable "region" {
  description = "Region to launch the droplet. e.g. nyc1 (see https://slugs.do-api.dev/ for whole list)"
}

variable "public_key_path" {
  description = "The absolute path of the public key used to log into the server"
}

