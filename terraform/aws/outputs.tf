output "ip" {
  value = aws_instance.minecraft.public_ip
}
output "permanent_device" {
  value = "/dev/nvme1n1"
}
