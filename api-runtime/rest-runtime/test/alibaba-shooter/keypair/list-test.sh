#!/bin/bash
source ../setup.env

for NAME in "${CONNECT_NAMES[@]}"
do
#	NAME=${CONNECT_NAMES[0]}
        curl -sX GET http://$RESTSERVER:1024/spider/keypair -H 'Content-Type: application/json' -d '{ "ConnectionName": "'${NAME}'" }' |json_pp &
done
