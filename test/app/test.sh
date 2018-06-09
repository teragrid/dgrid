#! /bin/bash
set -ex

#- kvstore over socket, curl
#- counter over socket, curl
#- counter over grpc, curl
#- counter over grpc, grpc

# TODO: install everything

export PATH="$GOBIN:$PATH"
export TMHOME=$HOME/.teragrid_app

function kvstore_over_socket(){
    rm -rf $TMHOME
    teragrid init
    echo "Starting kvstore_over_socket"
    asura-cli kvstore > /dev/null &
    pid_kvstore=$!
    teragrid node > teragrid.log &
    pid_teragrid=$!
    sleep 5

    echo "running test"
    bash kvstore_test.sh "KVStore over Socket"

    kill -9 $pid_kvstore $pid_teragrid
}

# start teragrid first
function kvstore_over_socket_reorder(){
    rm -rf $TMHOME
    teragrid init
    echo "Starting kvstore_over_socket_reorder (ie. start teragrid first)"
    teragrid node > teragrid.log &
    pid_teragrid=$!
    sleep 2
    asura-cli kvstore > /dev/null &
    pid_kvstore=$!
    sleep 5

    echo "running test"
    bash kvstore_test.sh "KVStore over Socket"

    kill -9 $pid_kvstore $pid_teragrid
}


function counter_over_socket() {
    rm -rf $TMHOME
    teragrid init
    echo "Starting counter_over_socket"
    asura-cli counter --serial > /dev/null &
    pid_counter=$!
    teragrid node > teragrid.log &
    pid_teragrid=$!
    sleep 5

    echo "running test"
    bash counter_test.sh "Counter over Socket"

    kill -9 $pid_counter $pid_teragrid
}

function counter_over_grpc() {
    rm -rf $TMHOME
    teragrid init
    echo "Starting counter_over_grpc"
    asura-cli counter --serial --asura grpc > /dev/null &
    pid_counter=$!
    teragrid node --asura grpc > teragrid.log &
    pid_teragrid=$!
    sleep 5

    echo "running test"
    bash counter_test.sh "Counter over GRPC"

    kill -9 $pid_counter $pid_teragrid
}

function counter_over_grpc_grpc() {
    rm -rf $TMHOME
    teragrid init
    echo "Starting counter_over_grpc_grpc (ie. with grpc broadcast_tx)"
    asura-cli counter --serial --asura grpc > /dev/null &
    pid_counter=$!
    sleep 1
    GRPC_PORT=36656
    teragrid node --asura grpc --rpc.grpc_laddr tcp://localhost:$GRPC_PORT > teragrid.log &
    pid_teragrid=$!
    sleep 5

    echo "running test"
    GRPC_BROADCAST_TX=true bash counter_test.sh "Counter over GRPC via GRPC BroadcastTx"

    kill -9 $pid_counter $pid_teragrid
}

cd $GOPATH/src/github.com/teragrid/teragrid/test/app

case "$1" in 
    "kvstore_over_socket")
    kvstore_over_socket
    ;;
"kvstore_over_socket_reorder")
    kvstore_over_socket_reorder
    ;;
    "counter_over_socket")
    counter_over_socket
    ;;
"counter_over_grpc")
    counter_over_grpc
    ;;
    "counter_over_grpc_grpc")
    counter_over_grpc_grpc
    ;;
*)
    echo "Running all"
    kvstore_over_socket
    echo ""
    kvstore_over_socket_reorder
    echo ""
    counter_over_socket
    echo ""
    counter_over_grpc
    echo ""
    counter_over_grpc_grpc
esac

