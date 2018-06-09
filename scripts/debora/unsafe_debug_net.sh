#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; killall teragrid"
debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; teragrid unsafe_reset_priv_validator; rm -rf ~/.teragrid/data"
debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; git pull origin develop; make"
debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; mkdir -p ~/.teragrid/logs"
debora run --bg --label teragrid -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; teragrid node 2>&1 | stdinwriter -outpath ~/.teragrid/logs/teragrid.log"
printf "\n\nSleeping for a minute\n"
sleep 60
debora download teragrid "logs/async$1"
debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; killall teragrid"
