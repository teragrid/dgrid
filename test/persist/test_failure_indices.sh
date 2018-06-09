#! /bin/bash

export PATH="$GOBIN:$PATH"
export TMHOME=$HOME/.teragrid_persist

rm -rf "$TMHOME"
teragrid init

# use a unix socket so we can remove it
RPC_ADDR="$(pwd)/rpc.sock"

TM_CMD="teragrid node --log_level=debug --rpc.laddr=unix://$RPC_ADDR" # &> teragrid_${name}.log"
DUMMY_CMD="asura-cli kvstore --persist $TMHOME/kvstore" # &> kvstore_${name}.log"


function start_procs(){
    name=$1
    indexToFail=$2
    echo "Starting persistent kvstore and teragrid"
    if [[ "$CIRCLECI" == true ]]; then
        $DUMMY_CMD &
    else
        $DUMMY_CMD &> "kvstore_${name}.log" &
    fi
    PID_DUMMY=$!

    # before starting teragrid, remove the rpc socket
    rm -f $RPC_ADDR
    if [[ "$indexToFail" == "" ]]; then
        # run in background, dont fail
		if [[ "$CIRCLECI" == true ]]; then
			$TM_CMD &
		else
            $TM_CMD &> "teragrid_${name}.log" &
		fi
        PID_teragrid=$!
    else
        # run in foreground, fail
		if [[ "$CIRCLECI" == true ]]; then
			FAIL_TEST_INDEX=$indexToFail $TM_CMD
		else
            FAIL_TEST_INDEX=$indexToFail $TM_CMD &> "teragrid_${name}.log"
		fi
        PID_teragrid=$!
    fi
}

function kill_procs(){
    kill -9 "$PID_DUMMY" "$PID_teragrid"
    wait "$PID_DUMMY"
    wait "$PID_teragrid"
}

# wait for port to be available
function wait_for_port() {
    port=$1
    # this will succeed while port is bound
    nc -z 127.0.0.1 $port
    ERR=$?
    i=0
    while [ "$ERR" == 0 ]; do
        echo "... port $port is still bound. waiting ..."
        sleep 1
        nc -z 127.0.0.1 $port
        ERR=$?
        i=$((i + 1))
        if [[ $i == 10 ]]; then
            echo "Timed out waiting for port to be released"
            exit 1
        fi
    done
    echo "... port $port is free!"
}


failsStart=0
fails=$(grep -r "fail.Fail" --include \*.go . | wc -l)
failsEnd=$((fails-1))

for failIndex in $(seq $failsStart $failsEnd); do
    echo ""
    echo "* Test FailIndex $failIndex"
    # test failure at failIndex

    bash ./test/utils/txs.sh "localhost:46657" &
    start_procs 1 "$failIndex"

    # teragrid should already have exited when it hits the fail index
    # but kill -9 for good measure
    kill_procs

    start_procs 2

    # wait for node to handshake and make a new block
    # NOTE: --unix-socket is only available in curl v7.40+
    curl -s --unix-socket "$RPC_ADDR" http://localhost/status > /dev/null
    ERR=$?
    i=0
    while [ "$ERR" != 0 ]; do
        sleep 1
        curl -s --unix-socket "$RPC_ADDR" http://localhost/status > /dev/null
        ERR=$?
        i=$((i + 1))
        if [[ $i == 20 ]]; then
            echo "Timed out waiting for teragrid to start"
            exit 1
        fi
    done

    # wait for a new block
    h1=$(curl -s --unix-socket "$RPC_ADDR" http://localhost/status | jq .result.sync_info.latest_block_height)
    h2=$h1
    while [ "$h2" == "$h1" ]; do
        sleep 1
        h2=$(curl -s --unix-socket "$RPC_ADDR" http://localhost/status | jq .result.sync_info.latest_block_height)
    done

    kill_procs

    echo "* Passed Test for FailIndex $failIndex"
    echo ""
done

echo "Passed Test: Persistence"
