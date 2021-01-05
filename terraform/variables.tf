variable "do_token" {
  default = ""
}

variable "cloud_provider" {
  description = "The cloud provider used. Currently supports AWS and Digital ocean."
}

variable "server_name" {
  description = "The unique server ID (I put it as guild name)"
}

variable "public_key_path" {
  description = "The absolute path of the public key used to log into the server"
}

