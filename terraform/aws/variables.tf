
variable "user_data" {
  description = "cloudinit user data"
}

variable "server_name" {
  description = "human readable name of server"
}

variable "instance_size" {
  description = "AWS ec2 instance type. e.g. t3.micro (see https://www.ec2instances.info/ for whole list)"
}

variable "region" {
  description = "Availability zone (e.g. us-east-1a)"
}
