#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

printf "Starting group $1...\n"
sleep 3

debora --group "$1" run --bg --label teragrid -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; teragrid node 2>&1 | stdinwriter -outpath ~/.teragrid/logs/teragrid.log"
debora --group "$1" run --bg --label logjack    -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; logjack -chopSize='10M' -limitSize='1G' ~/.teragrid/logs/teragrid.log"
printf "Done\n"
