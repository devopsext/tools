#!/bin/bash

set -e

pushd `dirname $0` > /dev/null
SCRIPTPATH=`pwd -P`
popd > /dev/null
SCRIPTFILE=`basename $0`

mkdir -p ${SCRIPTPATH}/.certs

cd ${SCRIPTPATH}/.certs

DAYS=3650
SIZE=4096
CN="tools"
DNS="tools"

# generate a self-signed rootCA file that would be used to sign both the server and client cert.
# Alternatively, we can use different CA files to sign the server and client, but for our use case, we would use a single CA.

echo "Creating new key and certificate for Root CA..."
openssl req -newkey rsa:${SIZE} \
  -new -nodes -x509 \
  -days ${DAYS} \
  -out ca.crt \
  -keyout ca.key \
  -subj "/CN=${CN}"

#create a key for server

echo "Creating new key and certificate for Server..."
openssl genrsa -out server.key ${SIZE}

#generate the Certificate Signing Request

echo "Creating new CSR for Server..."
openssl req -new -key server.key -out server.csr \
  -subj "/CN=${CN}"

#sign it with Root CA
# https://stackoverflow.com/questions/64814173/how-do-i-use-sans-with-openssl-instead-of-common-name

echo "Signing the Server certificate with Root CA..."
openssl x509  -req -in server.csr \
  -extfile <(printf "subjectAltName=DNS:${DNS}") \
  -CA ca.crt -CAkey ca.key  \
  -days ${DAYS} -sha256 -CAcreateserial \
  -out server.crt

cat server.crt server.key > server.pem

function generate_client() {
  CLIENT=$1
  O=$2
  OU=$3

  echo "Creating new key and certificate for ${CLIENT}..."
  openssl genrsa -out ${CLIENT}.key ${SIZE}

  echo "Creating new CSR for ${CLIENT}..."
  openssl req -new -key ${CLIENT}.key -out ${CLIENT}.csr \
    -subj "/CN=${CN}"

  echo "Signing the ${CLIENT} certificate with Root CA..."
  openssl x509  -req -in ${CLIENT}.csr \
    -extfile <(printf "subjectAltName=DNS:${DNS}") \
    -CA ca.crt -CAkey ca.key -out ${CLIENT}.crt -days ${DAYS} -sha256 -CAcreateserial

  cat ${CLIENT}.crt ${CLIENT}.key > ${CLIENT}.pem
}

generate_client client
