#!/bin/sh
SCRIPT=$(readlink -f "$0")
BASEDIR=$(dirname "$SCRIPT")
cd $BASEDIR || exit 1

echo "commons" 
cd $BASEDIR/commons/ts && npm ci --force --verbose && npm run build --verbose && \
echo "admin-ui" && \
cd $BASEDIR/admin-ui && npm ci --force --verbose && REACT_APP_PRODUCT_VERSION=$(cat ../version.txt | awk NF) npm run build  --verbose && \
echo "booking-ui" && \
cd $BASEDIR/booking-ui && npm ci --force --verbose && REACT_APP_PRODUCT_VERSION=$(cat ../version.txt | awk NF) npm run build --verbose && \
cd $BASEDIR && \
echo DONE

