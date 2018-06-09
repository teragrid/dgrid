#! /bin/bash

# get the asura commit used by teragrid
COMMIT=$(bash scripts/dep_utils/parse.sh asura)
echo "Checking out vendored commit for asura: $COMMIT"

go get -d github.com/teragrid/asura
cd "$GOPATH/src/github.com/teragrid/asura" || exit
git checkout "$COMMIT"
make get_tools
make get_vendor_deps
make install
