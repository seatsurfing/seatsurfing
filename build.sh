#!/bin/sh
SCRIPT=$(readlink -f "$0")
BASEDIR=$(dirname "$SCRIPT")
cd $BASEDIR || exit 1

echo "booking-ui" && \
cd $BASEDIR/booking-ui && npm ci --force --verbose && REACT_APP_PRODUCT_VERSION=$(cat ../version.txt | awk NF) npm run build --verbose && \
cd $BASEDIR && \
echo DONE

