#!/bin/bash

if [ "$1" = "" ]; then
        echo
        echo -e 'usage: '$0' mock|aws|azure|gcp|alibaba|tencent|ibm|openstack|ncp|nhncloud|ncpvpc|ktvpc'
        echo -e '\n\tex) '$0' aws'
        echo
        exit 0;
fi

source $1/setup.env

PRTL1=TCP
PORT1=23

PRTL2=TCP
PORT2=22

PRTL3=TCP
PORT3=22

INTERVAL=10
TIMEOUT=9
THRESHOLD=3

./common/create-nlb-test.sh $PRTL1 $PORT1 $PRTL2 $PORT2 $PRTL3 $PORT3 $INTERVAL $TIMEOUT $THRESHOLD
