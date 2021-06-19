require('./sni_pb');
const services = require('./sni_grpc_pb');
const grpc = require('@grpc/grpc-js');

function main() {
  const target = 'localhost:8191';

  const client = new services.DevicesClient(target, grpc.credentials.createInsecure());

  const request = new proto.DevicesRequest();
  //request.addKinds("retroarch");

  client.listDevices(request, function(err, response) {
    if (err) {
      console.error(err);
    }
    if (response) {
      console.log('Devices:');
      for (let dev of response.getDevicesList()) {
        const uri = dev.getUri();
        const displayName = dev.getDisplayname();
        const kind = dev.getKind();
        const caps = dev.getCapabilitiesList();
        console.log(`  uri:         ${uri}`);
        console.log(`  displayName: ${displayName}`);
        console.log(`  kind:        ${kind}`);
        console.log(`  caps:        ${caps}`);
      }
    }
  });
}

main();
