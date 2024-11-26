#!/bin/bash

CSPLIST=( aws azure gcp alibaba tencent ibm openstack ktcloud ktcloudvpc ncp ncpvpc nhncloud )

function run() {
    param=$1
    num=0
    for CSP in "${CSPLIST[@]}"
    do
        echo "============ test ${CSP} ... ============"

	./one_csp_run.sh ${CSP} ${param} &

        echo -e "\n\n"
    done
}

run "$@"

