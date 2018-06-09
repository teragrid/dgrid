#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

debora open "[::]:46661"
debora --group default.upgrade status
printf "\n\nShutting down barak default port...\n\n"
sleep 3
debora --group default.upgrade close "[::]:46660"
debora --group default.upgrade run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; git pull origin develop; make"
debora --group default.upgrade run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; mkdir -p ~/.barak/logs"
debora --group default.upgrade run --bg --label barak -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; barak --config=cmd/barak/seed 2>&1 | stdinwriter -outpath ~/.barak/logs/barak.log"
printf "\n\nTesting new barak...\n\n"
sleep 3
debora status
printf "\n\nShutting down old barak...\n\n"
sleep 3
debora --group default.upgrade quit
printf "Done!\n"
