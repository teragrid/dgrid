#! /bin/bash
set -eu

ID=$1
DOCKER_IMAGE=$2
NODEID="$(docker run --rm -e TMHOME=/go/src/github.com/teragrid/teragrid/test/p2p/data/mach$ID/core $DOCKER_IMAGE teragrid show_node_id)"
echo "$NODEID@172.57.0.$((100+$ID))"
