#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; killall teragrid; killall logjack"
debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; teragrid unsafe_reset_priv_validator; rm -rf ~/.teragrid/data; rm ~/.teragrid/config/genesis.json; rm ~/.teragrid/logs/*"
debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; git pull origin develop; make"
debora run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; mkdir -p ~/.teragrid/logs"
debora run --bg --label teragrid -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; teragrid node 2>&1 | stdinwriter -outpath ~/.teragrid/logs/teragrid.log"
debora run --bg --label logjack    -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; logjack -chopSize='10M' -limitSize='1G' ~/.teragrid/logs/teragrid.log"
printf "Done\n"
