#! /bin/bash

# update the `tester` image by copying in the latest teragrid binary

docker run --name builder tester true
docker cp $GOPATH/bin/teragrid builder:/go/bin/teragrid
docker commit builder tester
docker rm -vf builder

