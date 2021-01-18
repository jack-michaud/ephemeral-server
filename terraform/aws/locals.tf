
locals {
  user_data = templatefile("${path.module}/user_data.tmpl", {
    public_key_data = file(var.public_key_path)
  })
  // oh man, limitation. 
  // TODO Specify availability_zone, or dynamically pick the first one
  availability_zone = "${var.region}a"
}

