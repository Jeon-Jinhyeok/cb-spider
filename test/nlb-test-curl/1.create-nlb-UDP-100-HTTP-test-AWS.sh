#!/bin/bash

if [ "$1" = "" ]; then
        echo
        echo -e 'usage: '$0' mock|aws|azure|gcp|alibaba|tencent|ibm|openstack|ncp|nhncloud'
        echo -e '\n\tex) '$0' aws'
        echo
        exit 0;
fi

source $1/setup.env

PRTL1=UDP
PORT1=100

PRTL2=UDP
PORT2=100

PRTL3=HTTP
PORT3=80

INTERVAL=10
TIMEOUT=-1
THRESHOLD=3

./common/create-nlb-test.sh $PRTL1 $PORT1 $PRTL2 $PORT2 $PRTL3 $PORT3 $INTERVAL $TIMEOUT $THRESHOLD
