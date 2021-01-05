#!/bin/bash
# Author: @jack-michaud

base_dir=$(dirname $0)
terraform_dir=$base_dir/terraform
config_dir=$base_dir/config
ansible_dir=$base_dir/ansible
key_dir=$config_dir/keys

function build_env {
    source .env
    if [ -z $DIGITAL_OCEAN_TOKEN ]; then
      echo 'Error: must provide DIGITAL_OCEAN_TOKEN in environment variables (or in .env)'
      exit 1
    fi
    if [ -z $DIGITAL_OCEAN_TOKEN ]; then
      echo 'Error: must provide DIGITAL_OCEAN_TOKEN in environment variables (or in .env)'
      exit 1
    fi
}

function generate_keypair {
  keyname=minecraft
  [ -d $key_dir ] || mkdir -p $key_dir

  # Generate ssh keys if the keys do not exist
  [ ! -f $key_dir/$keyname.pub ] || [ ! -f $key_dir/$keyname ] && (
    rm $key_dir/$keyname*
    ssh-keygen -q -N '' -f $key_dir/$keyname 
    #ssh-keygen -b 521 -t ecdsa -N '' -f $key_dir/$keyname 
  )

  echo $(realpath $key_dir/$keyname)
}

function initialize {
  [[ -d $terraform_dir/.terraform ]] || terraform -chdir=$terraform_dir init
  generate_keypair > /dev/null
}

function apply_terraform {
    PUBLIC_KEY_PATH=$(generate_keypair).pub
    pushd $terraform_dir > /dev/null
      #-var "do_token=$DIGITAL_OCEAN_TOKEN" \
    terraform \
      apply \
      -var "cloud_provider=aws" \
      -var "server_name=$SERVER_NAME" \
      -var "public_key_path=$PUBLIC_KEY_PATH" \
      -auto-approve
    popd > /dev/null
}

function destroy_server {
    PUBLIC_KEY_PATH=$(generate_keypair).pub
    pushd $terraform_dir > /dev/null
    terraform destroy \
      -var "cloud_provider=digitalocean" \
      -var "server_name=$SERVER_NAME" \
      -var "do_token=$DIGITAL_OCEAN_TOKEN" \
      -var "public_key_path=$PUBLIC_KEY_PATH" \
      -auto-approve \
      -target='module.digitalocean[0].digitalocean_droplet.minecraft' \
      -target='module.aws[0].aws_instance.minecraft'
    popd > /dev/null
}
function destroy_all {
    PUBLIC_KEY_PATH=$(generate_keypair).pub
    pushd $terraform_dir > /dev/null
    terraform destroy \
      -var "cloud_provider=digitalocean" \
      -var "server_name=$SERVER_NAME" \
      -var "do_token=$DIGITAL_OCEAN_TOKEN" \
      -var "public_key_path=$PUBLIC_KEY_PATH" \
      -auto-approve
    popd > /dev/null
}

function get_ip {
    pushd $terraform_dir > /dev/null
    terraform output | grep 'ip' | awk '{ print $3 }'
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
  ansible-playbook -i $config_dir/ansible_inventory -e server_type=$SERVER_TYPE -e persistent_device=$DEVICE $ansible_dir/main.yml
}

initialize

while getopts "dDciIn:t:" OPTION; do
  case $OPTION in
    D) ACTION='destroy_all' ;;
    d) ACTION='destroy' ;;
    c) ACTION='create' ;;
    i) ACTION='get_ip' ;;
    I) ACTION='ansible_install' ;;
    n) SERVER_NAME=$OPTARG ;;
    t) SERVER_TYPE=$OPTARG ;;
  esac
done

if [ -z $SERVER_NAME ]; then
  echo 'Must supply -n <server name> option.'
  exit 1
fi

if [ -z $SERVER_TYPE ]; then
  SERVER_TYPE=texkit3
fi


case $ACTION in
  'destroy_all')
    build_env && destroy_all
    exit 0
    ;;
  'destroy')
    build_env && destroy_server
    exit 0
    ;;
  'create')
    build_env && apply_terraform && ansible_install
    exit 0
    ;;
  'get_ip')
    build_env && get_ip
    exit 0
    ;;
  'ansible_install')
    build_env && ansible_install
    exit 0
    ;;
  *)
    echo 'Invalid option! Must specify n (server name) and t (server type) and one of d (destroy), D (destroy all), c (create), i (get IP)'
    exit 1
    ;;
esac

echo 'No option specified'
