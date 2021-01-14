
locals {
  user_data = templatefile("${path.module}/user_data.tmpl", {
    public_key_data = file(var.public_key_path)
  })
}

