import {grpc} from "@improbable-eng/grpc-web";

// Import code-generated data structures.
import {Devices, DevicesClient} from "../sni-client/sni_pb_service";
import {DevicesRequest, DevicesResponse} from "../sni-client/sni_pb";

const host = "http://localhost:8190";

const req = new DevicesRequest();
grpc.unary(Devices.ListDevices, {
    request: req,
    host: host,
    onEnd: res => {
        const { status, statusMessage, headers, message, trailers } = res;
        if (status === grpc.Code.OK && message) {
            console.log("all ok. got devices: ", message.toObject());
        }
    }
});
