#!/bin/sh
# The Go code for sni.proto lives in the snes-devices repository (github.com/alttpo/snes-devices/protos/sni) so that
# SNI and the snes-devices drivers share a single registration of sni.proto. This script expects a snes-devices
# checkout next to the sni checkout; keep the two copies of sni.proto in sync.
set -e
DEST=../../../snes-devices/protos/sni
cp sni.proto "$DEST/"
cd "$DEST"
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    sni.proto
