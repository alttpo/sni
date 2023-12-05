get_protoc_filename() {
  case "$(uname -s)" in
    CYGWIN*|MINGW*|MSYS*) # Check if running in Windows
      echo "node_modules\\.bin\\protoc-gen-ts.cmd" ;;
    *)
      echo "node_modules/.bin/protoc-gen-ts" ;;
  esac
}

filename=$(get_protoc_filename)
protoc \
--plugin=protoc-gen-ts=$filename \
--js_out=import_style=commonjs,binary:sni-client \
--ts_out=service=grpc-web:sni-client \
-I ../../protos/sni/ \
../../protos/sni/*.proto
