set -e

[ -d sni ] || mkdir sni

protoc --proto_path=../../protos/sni --go_out=./sni --go-grpc_out=./sni --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ../../protos/sni/sni.proto
