#!/bin/bash

function generateCertificates() {

  local NAME="$1"
  local DAYS="$2"

  DAYS=${DAYS:-365}

  if [[ "$NAME" == "" ]]; then
    echo "Name is not specified. Generation skipped."
    return
  fi

  local KEY_FILE="$NAME.key"
  local CRT_FILE="$NAME.crt"

  if [[ "$KEY_FILE" == "" ]] || [[ "$CRT_FILE" == "" ]]; then
    echo "Key/Crt file name is not specified. Generation skipped."
    return
  fi

  tmpdir=$(mktemp -d)

  cat <<EOF >>"${tmpdir}/csr.conf"
  [req]
  req_extensions = v3_req
  distinguished_name = req_distinguished_name
  [req_distinguished_name]
  [ v3_req ]
  basicConstraints = CA:FALSE
  keyUsage = nonRepudiation, digitalSignature, keyEncipherment
  extendedKeyUsage = serverAuth
  subjectAltName = @alt_names
  [alt_names]
  DNS.1 = ${NAME}
EOF

  local __out=""

  openssl genrsa -out "${KEY_FILE}" 4096 >__openSsl.out 2>&1
  __out=$(cat __openSsl.out)
  if [[ ! "$?" -eq 0 ]]; then
    echo "$__out"
    return 1
  else
    echo "'openssl genrsa' output:\n$__out"
  fi

  openssl req -new -key "${KEY_FILE}" -subj "/CN=${NAME}" -out "${tmpdir}/${NAME}.csr" -config "${tmpdir}/csr.conf" >__openSsl.out 2>&1
  __out=$(cat __openSsl.out)
  if [[ ! "$?" -eq 0 ]]; then
    echo "$__out"
    return 1
  else
    echo "'openssl req -new -key' output:\n$__out"
  fi

  openssl x509 -signkey "${KEY_FILE}" -in "${tmpdir}/${NAME}.csr" -req -days $DAYS -out "${CRT_FILE}" >__openSsl.out 2>&1
  __out=$(cat __openSsl.out)
  if [[ ! "$?" -eq 0 ]]; then
    echo "$__out"
    return 1
  else
    echo "'openssl x509 -signkey' output:\n$__out"
  fi

  rm -rf __openSsl.out
}

generateCertificates "$1" "$2"

echo "Crt..."
cat "$1.crt" | base64 | tr -d '\n'
echo ""
echo "Key..."
cat "$1.key" | base64 | tr -d '\n'
echo ""
echo "Bundle..."
cat "$1.crt" | base64 | tr -d '\n'
echo ""% 