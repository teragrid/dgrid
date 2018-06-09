#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

printf "Upgrading group $1...\n"
sleep 3

debora --group "$1" run -- bash -c "cd \$GOPATH/src/github.com/teragrid/teragrid; git pull origin develop; make"
printf "Done\n"
