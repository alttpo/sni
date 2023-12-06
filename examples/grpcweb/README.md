# SNI for Web

## Requirements
* [Google Protocol Buffers](https://grpc.io/docs/protoc-installation/) (ie `protoc`)
* Node 18+

## Setup
First, you will need to install the Node.js dependencies.
```sh
npm install
```

Once that is done installing, you can generate the client files with the `build` script.
```sh
npm run build
```

Alternitavely, you can manually call the `gen.sh` script.
```sh
./gen.sh
```

This will populate the `sni-client` folder with Javascript and Typescript definitions to use in your project.
