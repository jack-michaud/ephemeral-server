resource "aws_ebs_volume" "mc_vol" {
  tags = {
    name = "minecraft-volume-${var.server_name}"
  }
  availability_zone = var.region 
  size = 20
}


data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_security_group" "minecraft_security_group" {
  // We use a host based firewall controlled in ansible
  name = "allow-all-${var.server_name}"
  description = "Allow all traffic inbound and outbound (use Host based firewall!)"

  ingress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = {
    Name = "minecraft-allow-all"
  }
}

resource "aws_instance" "minecraft" {
  ami = data.aws_ami.ubuntu.id
  tags = {
    name   = "minecraft-server-${var.server_name}"
  }
  availability_zone = var.region
  instance_type   = var.instance_size

  user_data = local.user_data

  security_groups = [
    aws_security_group.minecraft_security_group.name
  ]
}

resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sda2"
  volume_id   = aws_ebs_volume.mc_vol.id
  instance_id = aws_instance.minecraft.id
}

