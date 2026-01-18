#!/bin/bash
set -e

HOSTS=("mongodb1" "mongodb2" "mongodb3" "mongodb4" "mongo-arbiter")
PORT=27017
ARBITER_PORT=27018

USER="${MONGO_INITDB_ROOT_USERNAME}"
PASS="${MONGO_INITDB_ROOT_PASSWORD}"
RS_NAME="rs0"
TIMEOUT=10

echo "Waiting for MongoDB nodes to respond..."

for host in "${HOSTS[@]}"; do
    if [[ "$host" == "mongo-arbiter" ]]; then
        port=$ARBITER_PORT
    else
        port=$PORT
    fi

    until mongosh --host "$host" --port "$port" --quiet \
        --eval "db.runCommand({ ping: 1 }).ok" &>/dev/null; do
        echo -n "."
        sleep 1
    done

    echo " $host ready"
done


echo "Checking if replica set already initialized..."

if mongosh --host "${HOSTS[0]}" \
    -u "$USER" -p "$PASS" --authenticationDatabase admin \
    --quiet --eval "rs.status().ok" &>/dev/null; then
    echo "Replica set already initialized"
    exit 0
fi

echo "Building replica set config..."

CONFIG=$(cat <<EOF
{
  "_id": "$RS_NAME",
  "members": [
    { "_id": 0, "host": "${HOSTS[0]}:$PORT", "priority": 5 },
    { "_id": 1, "host": "${HOSTS[1]}:$PORT", "priority": 4 },
    { "_id": 2, "host": "${HOSTS[2]}:$PORT", "priority": 3 },
    { "_id": 3, "host": "${HOSTS[3]}:$PORT", "priority": 2 },

    { "_id": 4, "host": "${HOSTS[4]}:${ARBITER_PORT}", "arbiterOnly": true}
  ]
}
EOF
)

echo "Applying replica set configuration..."

mongosh --host "${HOSTS[0]}" \
    -u "$USER" -p "$PASS" --authenticationDatabase admin <<EOF
try {
    rs.initiate($CONFIG);
} catch (e) {
    if (e.codeName === "AlreadyInitialized") {
        print("Replica set already exists, reconfiguring...");
        rs.reconfig($CONFIG, { force: true });
    } else {
        throw e;
    }
}

print("Waiting for PRIMARY election...");
var start = Date.now();
while (true) {
    var hello = db.hello();
    if (hello.isWritablePrimary) {
        print("PRIMARY elected: " + hello.me);
        break;
    }
    if (Date.now() - start > ${TIMEOUT} * 1000) {
        print("Timeout waiting for PRIMARY");
        quit(1);
    }
    sleep(1000);
}

printjson(rs.status());
EOF

exit $?
