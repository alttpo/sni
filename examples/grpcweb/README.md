# SNI gRPC for Web

## Requirements
* Node 18+

## Setup
First, you will need to install the Node.js dependencies.
```sh
npm install
```

Once that is done installing, you can run the page locally by using the `dev` command.
```sh
npm run dev
```
You can now visit the project at [`localhost:3000`](http://localhost:3000) and list your connected devices.

## Generating Client Files
You can generate the client files independently with the `compile` command. This is done automatically in `dev` and `build`.
```sh
npm run compile
```

This will populate the `lib` folder with Javascript and Typescript definitions to use in your project.
