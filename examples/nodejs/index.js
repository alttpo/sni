require('./sni_pb')
const services = require('./sni_grpc_pb');
const grpc = require('@grpc/grpc-js');

function promisify(c) {
  return new Promise((resolve, reject) => {
    c((err, rsp) => {
      if (err) reject(err);
      else if (rsp) resolve(rsp);
    });
  });
}

async function main() {
  const target = 'localhost:8191';

  const client = new services.DevicesClient(target, grpc.credentials.createInsecure());

  async function getDevices() {
    const req = new sni.DevicesRequest();

    //request.addKinds("retroarch");

    return await promisify(client.listDevices.bind(client, req));
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

  if (devices.length > 0) {
    const memory = new services.DeviceMemoryClient(target, grpc.credentials.createInsecure());

    let mapping = sni.MemoryMapping.LOROM;
    {
      const d = new sni.DetectMemoryMappingRequest();
      d.setUri(devices[0].getUri());
      d.setFallbackmemorymapping(sni.MemoryMapping.LOROM);
      const detectRsp = await promisify(memory.mappingDetect.bind(memory, d));
      mapping = detectRsp.getMemorymapping();
    }

    {
      const r = new sni.SingleReadMemoryRequest();
      r.setUri(devices[0].getUri());
      {
        const rr = new sni.ReadMemoryRequest();
        rr.setRequestaddress(0x7E0010);
        rr.setRequestaddressspace(sni.AddressSpace.SNESABUS);
        rr.setRequestmemorymapping(mapping);
        rr.setSize(1);
        r.setRequest(rr);
      }

      for (let i = 0; i < 240; i++) {
        const readRsp = await promisify(memory.singleRead.bind(memory, r));

        // console.log(readRsp.getResponse().getData_asB64());
        console.log(readRsp.getResponse().getData()[0]);
      }
    }
  }
}

main().then(_ => {
});
