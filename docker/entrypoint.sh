#!/bin/sh
mkdir -p /opt/keystore/
cat /vault/secrets/keystore | jq -c -r .data.keystoreValidator01 > /opt/keystore/keystoreValidator01
cat /vault/secrets/keystore | jq -c -r .data.keystoreValidator02 > /opt/keystore/keystoreValidator02
exec bridge --config $CONFIG_PATH --verbosity $VERBOSITY
