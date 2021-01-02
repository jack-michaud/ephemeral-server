#!/bin/bash
# Author: @jack-michaud

base_dir=$(dirname $0)
terraform_dir=$base_dir/terraform
key_dir=$base_dir/config/keys

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

  cat $key_dir/$keyname.pub
}

function initialize {
  [[ -d $terraform_dir/.terraform ]] || terraform -chdir=$terraform_dir init
  generate_keypair
}

function apply_terraform {
    PUBLIC_KEY=$(generate_keypair)
    pushd $terraform_dir
    terraform \
      apply \
      -var "server_name=$SERVER_NAME" \
      -var "server_type=$SERVER_TYPE" \
      -var "do_token=$DIGITAL_OCEAN_TOKEN" \
      -var "public_key=$PUBLIC_KEY" \
      -auto-approve
    popd
}

function destroy_server {
    PUBLIC_KEY=$(generate_keypair)
    pushd $terraform_dir
    terraform destroy \
      -var "server_name=$SERVER_NAME" \
      -var "server_type=$SERVER_TYPE" \
      -var "do_token=$DIGITAL_OCEAN_TOKEN" \
      -var "public_key=$PUBLIC_KEY" \
      -auto-approve \
      -target=digitalocean_droplet.minecraft
    popd
}
function destroy_all {
    pushd $terraform_dir
    terraform destroy \
      -var "server_name=$SERVER_NAME" \
      -var "server_type=$SERVER_TYPE" \
      -var "do_token=$DIGITAL_OCEAN_TOKEN" \
      -var "public_key=$PUBLIC_KEY" \
      -auto-approve
    popd
}

function get_ip {
    PUBLIC_KEY=$(generate_keypair)
    pushd $terraform_dir
    terraform output $base_dir/terraform | awk '{ print $3 }'
    popd
}

initialize

while getopts "dDcin:t:" OPTION; do
  case $OPTION in
    D) ACTION='destroy_all' ;;
    d) ACTION='destroy' ;;
    c) ACTION='create' ;;
    i) ACTION='get_ip' ;;
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
    build_env && apply_terraform
    exit 0
    ;;
  'get_ip')
    build_env && get_ip
    exit 0
    ;;
  *)
    echo 'Invalid option! Must specify n (server name) and t (server type) and one of d (destroy), D (destroy all), c (create), i (get IP)'
    exit 1
    ;;
esac

echo 'No option specified'
