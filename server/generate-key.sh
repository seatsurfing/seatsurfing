#!/bin/bash
set -e

OUT_DIR="${1:-.}"

openssl genrsa -traditional -out "$OUT_DIR/private.pem" 2048
openssl rsa -in "$OUT_DIR/private.pem" -pubout -out "$OUT_DIR/public.pem"

echo "Keys generated:"
echo "  Private key: $OUT_DIR/private.pem"
echo "  Public key:  $OUT_DIR/public.pem"
echo ""
echo "Set the following environment variables:"
echo "  JWT_PRIVATE_KEY=private.pem"
echo "  JWT_PUBLIC_KEY=public.pem"
