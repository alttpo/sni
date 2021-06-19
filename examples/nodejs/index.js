const messages = require('./sni_pb');
const services = require('./sni_grpc_pb');
const grpc = require('@grpc/grpc-js');

async function main() {
  const target = 'localhost:8191';

  const client = new services.DevicesClient(target, grpc.credentials.createInsecure());

  async function getDevices() {
    const req = new messages.DevicesRequest();

    //request.addKinds("retroarch");

    return await new Promise((resolve, reject) => {
      client.listDevices(req, (err, rsp) => {
        if (err) reject(err);
        else resolve(rsp);
      });
    });
  }

  const devices = (await getDevices()).getDevicesList();
  console.log('Devices:');
  for (let dev of devices) {
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

main().then(_ => {});
