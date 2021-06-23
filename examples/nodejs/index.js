/* eslint-disable */
// @ts-nocheck

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

function promisifyObj(c) {
  return new Promise((resolve, reject) => {
    c((err, rsp) => {
      if (err) reject(err);
      else if (rsp) resolve(rsp.toObject());
    });
  });
}

async function main() {
  const target = 'localhost:8191';

  const client = new services.DevicesClient(target, grpc.credentials.createInsecure());

  async function getDevices() {
    const req = new sni.DevicesRequest();

    //request.addKinds("retroarch");

    return await promisifyObj(client.listDevices.bind(client, req));
  }

  const devicesResponse = await getDevices();
  const devicesList = devicesResponse.devicesList;
  console.log('Devices:');
  for (let dev of devicesList) {
    console.log(`  uri:         ${(dev.uri)}`);
    console.log(`  displayName: ${(dev.displayname)}`);
    console.log(`  kind:        ${(dev.kind)}`);
    console.log(`  caps:        ${(dev.capabilitiesList)}`);
  }

  if (devicesList.length > 0) {
    const memory = new services.DeviceMemoryClient(target, grpc.credentials.createInsecure());

    let mapping = sni.MemoryMapping.LOROM;
    {
      const d = new sni.DetectMemoryMappingRequest();
      d.setUri(devicesList[0].uri);
      d.setFallbackmemorymapping(sni.MemoryMapping.LOROM);
      const detectRsp = await promisifyObj(memory.mappingDetect.bind(memory, d));
      mapping = detectRsp.memorymapping;
    }

    {
      const r = new sni.SingleReadMemoryRequest();
      r.setUri(devicesList[0].uri);
      {
        const rr = new sni.ReadMemoryRequest();
        rr.setRequestaddress(0x7E0010);
        rr.setRequestaddressspace(sni.AddressSpace.SNESABUS);
        rr.setRequestmemorymapping(mapping);
        rr.setSize(1);
        r.setRequest(rr);
      }

      for (let i = 0; i < 60*60; i++) {
        const readRsp = await promisifyObj(memory.singleRead.bind(memory, r));
        // FIXME: toObject() forcibly calls getData_asB64() instead of getData(), so we have to decode:
        const data = Buffer.from(readRsp.response.data, 'base64');
        console.log(data[0]);
      }
    }
  }
}

main().then(_ => {
});
