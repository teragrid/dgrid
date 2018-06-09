#! /bin/bash
set -e

function toHex() {
    echo -n $1 | hexdump -ve '1/1 "%.2X"' | awk '{print "0x" $0}' 

}

#####################
# kvstore with curl
#####################
TESTNAME=$1

# store key value pair
KEY="abcd"
VALUE="dcba"
echo $(toHex $KEY=$VALUE)
curl -s 127.0.0.1:46657/broadcast_tx_commit?tx=$(toHex $KEY=$VALUE)
echo $?
echo ""


###########################
# test using the asura-cli
###########################

echo "... testing query with asura-cli"

# we should be able to look up the key
RESPONSE=`asura-cli query \"$KEY\"`

set +e
A=`echo $RESPONSE | grep "$VALUE"`
if [[ $? != 0 ]]; then
    echo "Failed to find $VALUE for $KEY. Response:"
    echo "$RESPONSE"
    exit 1
fi
set -e

# we should not be able to look up the value
RESPONSE=`asura-cli query \"$VALUE\"`
set +e
A=`echo $RESPONSE | grep $VALUE`
if [[ $? == 0 ]]; then
    echo "Found '$VALUE' for $VALUE when we should not have. Response:"
    echo "$RESPONSE"
    exit 1
fi
set -e

#############################
# test using the /asura_query
#############################

echo "... testing query with /asura_query 2"

# we should be able to look up the key
RESPONSE=`curl -s "127.0.0.1:46657/asura_query?path=\"\"&data=$(toHex $KEY)&prove=false"`
RESPONSE=`echo $RESPONSE | jq .result.response.log`

set +e
A=`echo $RESPONSE | grep 'exists'`
if [[ $? != 0 ]]; then
    echo "Failed to find 'exists' for $KEY. Response:"
    echo "$RESPONSE"
    exit 1
fi
set -e

# we should not be able to look up the value
RESPONSE=`curl -s "127.0.0.1:46657/asura_query?path=\"\"&data=$(toHex $VALUE)&prove=false"`
RESPONSE=`echo $RESPONSE | jq .result.response.log`
set +e
A=`echo $RESPONSE | grep 'exists'`
if [[ $? == 0 ]]; then
    echo "Found 'exists' for $VALUE when we should not have. Response:"
    echo "$RESPONSE"
    exit 1
fi
set -e


echo "Passed Test: $TESTNAME"
