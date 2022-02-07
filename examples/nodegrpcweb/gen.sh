protoc \
--plugin=protoc-gen-ts=node_modules\\.bin\\protoc-gen-ts.cmd \
--js_out=import_style=commonjs,binary:sni-client \
--ts_out=service=grpc-web:sni-client \
-I ../../protos/sni/ \
../../protos/sni/*.proto
