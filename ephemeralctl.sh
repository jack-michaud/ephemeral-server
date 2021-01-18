#!/bin/bash
# Author: @jack-michaud

base_dir=$(realpath $(dirname $0))

function build_env {
    source .env
    if [ -z $CLOUD_PROVIDER ]; then
      echo 'Error: must provide CLOUD_PROVIDER in environment variables (or in .env)'
      exit 1
    fi
}

function generate_keypair {
  keyname=minecraft-$SERVER_NAME
  echo $(realpath $key_dir/$keyname)
}

function initialize {
  cd $base_dir
  if [[ "$CLOUD_PROVIDER" == "digitalocean" ]]; then
    if [ -z $DIGITALOCEAN_TOKEN ]; then
      echo 'Error: must provide DIGITALOCEAN_TOKEN in environment variables (or in .env)'
      exit 1
    fi
    src_terraform_dir=$base_dir/terraform/digitalocean
  fi
  if [[ "$CLOUD_PROVIDER" == "aws" ]]; then
    if [[ -z $AWS_ACCESS_KEY_ID ]]; then
      echo 'Error: must provide AWS_ACCESS_KEY_ID'
      exit 1
    fi
    if [[ -z $AWS_SECRET_ACCESS_KEY ]]; then
      echo 'Error: must provide AWS_SECRET_ACCESS_KEY'
      exit 1
    fi
    src_terraform_dir=$base_dir/terraform/aws
  fi
  config_dir=$base_dir/.cache/config-$SERVER_NAME
  ansible_dir=$base_dir/ansible
  key_dir=$config_dir/keys
  terraform_dir=$base_dir/.cache/terraform-$SERVER_NAME
  rm -r $terraform_dir
  mkdir -p $terraform_dir
  rsync -r $src_terraform_dir/* $terraform_dir
  pushd $terraform_dir > /dev/null
  cat <<EOF > terraform.tf
terraform {
  backend "consul" {
    address = "127.0.0.1:8500"
    scheme  = "http"
    path    = "tfstate/${SERVER_NAME}-server"
  }
}
EOF
  terraform init
  popd > /dev/null
  generate_keypair > /dev/null
}

function apply_terraform {
    PUBLIC_KEY_PATH=$(generate_keypair).pub
    pushd $terraform_dir > /dev/null

    terraform \
      apply \
      -var "region=$REGION" \
      -var "instance_size=$INSTANCE_SIZE" \
      -var "public_key_path=$PUBLIC_KEY_PATH" \
      -var "server_name=$SERVER_NAME" \
      -auto-approve
    status=$?

    popd > /dev/null
    test $status == 0 && echo 'Successfully applied terraform' || echo 'Failed to apply terraform'
    return $status
}

function destroy_server {
    PUBLIC_KEY_PATH=$(generate_keypair).pub
    pushd $terraform_dir > /dev/null
    terraform destroy \
      -var "region=$REGION" \
      -var "instance_size=$INSTANCE_SIZE" \
      -var "public_key_path=$PUBLIC_KEY_PATH" \
      -var "server_name=$SERVER_NAME" \
      -auto-approve \
      -target='module.digitalocean[0].digitalocean_droplet.minecraft' \
      -target='digitalocean_droplet.minecraft' \
      -target='aws_instance.minecraft' \
      -target='module.aws[0].aws_instance.minecraft'
    popd > /dev/null
}
function destroy_all {
    PUBLIC_KEY_PATH=$(generate_keypair).pub
    pushd $terraform_dir > /dev/null
    terraform destroy \
      -var "region=$REGION" \
      -var "instance_size=$INSTANCE_SIZE" \
      -var "public_key_path=$PUBLIC_KEY_PATH" \
      -var "server_name=$SERVER_NAME" \
      -auto-approve
    popd > /dev/null
}

function get_ip {
    pushd $terraform_dir > /dev/null
    terraform output | grep 'ip' | awk '{ print $3 }' | tr -d '"'
    popd > /dev/null
}

function get_device_location {
    pushd $terraform_dir > /dev/null
    terraform output | grep 'permanent_device' | awk '{ print $3 }'
    popd > /dev/null
}

function ansible_install {
  IP=$(get_ip)
  DEVICE=$(get_device_location)
  PRIVATE_KEY_FILE=$(generate_keypair)
  echo "minecraft ansible_host=${IP} ansible_user=minecraft ansible_port=22 ansible_ssh_private_key_file=${PRIVATE_KEY_FILE}" \
    > $config_dir/ansible_inventory

  ANSIBLE_HOST_KEY_CHECKING=False \
    ANSIBLE_SSH_RETRIES=5 \
    ansible-playbook -i $config_dir/ansible_inventory \
    -e server_type=$SERVER_TYPE \
    -e persistent_device=$DEVICE \
    $ansible_dir/main.yml
  status=$?

  test $status == 0 && echo 'Successfully applied ansible' || echo 'Failed to apply ansible'
  return $status
}


while getopts "dDciIs:t:n:r:" OPTION; do
  case $OPTION in
    D) ACTION='destroy_all' ;;
    d) ACTION='destroy' ;;
    c) ACTION='create' ;;
    i) ACTION='get_ip' ;;
    I) ACTION='ansible_install' ;;
    n) SERVER_NAME=$OPTARG ;;
    t) SERVER_TYPE=$OPTARG ;;
    s) INSTANCE_SIZE=$OPTARG ;;
    r) REGION=$OPTARG ;;
  esac
done

if [ -z $SERVER_NAME ]; then
  echo 'Must supply -n <server name> option.'
  exit 1
fi

if [ -z $INSTANCE_SIZE ]; then
  echo 'Must supply -s <instance size> option.'
  exit 1
fi


if [ -z $SERVER_TYPE ]; then
  echo 'Must supply -t server_type option.'
  exit 1
fi

build_env
initialize > /dev/null

case $ACTION in
  'destroy_all')
    destroy_all
    exit 0
    ;;
  'destroy')
    destroy_server
    exit 0
    ;;
  'create')
    (apply_terraform && ansible_install && exit 0) || exit 1
    ;;
  'get_ip')
    get_ip
    exit 0
    ;;
  'ansible_install')
    ansible_install
    exit 0
    ;;
  *)
    echo 'Invalid option! Must specify n (server name) and t (server type) and one of d (destroy), D (destroy all), c (create), i (get IP)'
    exit 1
    ;;
esac
