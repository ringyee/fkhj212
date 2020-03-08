#!/bin/bash
#########################################################################
# File Name: cmttest.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2020-03-02 12:42:59
#########################################################################

set -e

SEND_STR1=$(cat ./tfactors.json)
SEND_STR2=$(cat ./tupdataconf.json)
SEND_STR3=
SEND_STR4=
SEND_STR5=
CONFIG=
ADDR=127.0.0.1
PORT=5000
MN=010000A8900016F000169DCB
URL=http://${ADDR}:${PORT}/api/settings/$1/${MN}/$2
HEAD=Content-Type:application/json

if [[ x$1 == xr ]] && (( $2 < 6 && $2 >0 ));then 
    echo send get config command ......
    curl -i "${URL}" 
else
    if [[ x$1 == xw ]];then 
    echo send set config command ...... 
    eval "CONFIG=\$SEND_STR$2"
    echo CONFIG string = $CONFIG
    curl -X POST -H "${HEAD}" \
        -i "${URL}" \
        -d "$CONFIG"
    else
        echo "arguments [r,w] [1,2,3,4,5]" 
        exit 255
    fi
fi


exit $?
