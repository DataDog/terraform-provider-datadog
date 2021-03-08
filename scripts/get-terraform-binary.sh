#!/usr/bin/env bash
# Get terraform binary to use for running tests

VERSION=${TF_VERSION:?TF_VERSION must be provided.}
FILE_NAME=''

echo "OS is $(uname -s)"

case "$(uname -s)" in
   MINGW64*)
    wget -O tf.zip https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_windows_amd64.zip
    FILE_NAME=$(unzip -Z -1 tf.zip)
     ;;

   *)
    wget -O tf.zip https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_linux_amd64.zip
    FILE_NAME=$(unzip -Z -1 tf.zip)
     ;;
esac

unzip 'tf.zip'
TF_BINARY_PATH=$(pwd)/${FILE_NAME}

export TF_ACC_TERRAFORM_PATH=${TF_BINARY_PATH}
echo "TF binary path is: ${TF_ACC_TERRAFORM_PATH}"
