grpc_tools_node_protoc --proto_path=../../protos/sni --js_out=import_style=commonjs,namespace_prefix=sni,binary:./ --grpc_out=grpc_js:. ../../protos/sni/sni.proto
