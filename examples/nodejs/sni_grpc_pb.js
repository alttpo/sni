// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var sni_pb = require('./sni_pb.js');

function serialize_DetectMemoryMappingRequest(arg) {
  if (!(arg instanceof sni_pb.DetectMemoryMappingRequest)) {
    throw new Error('Expected argument of type DetectMemoryMappingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_DetectMemoryMappingRequest(buffer_arg) {
  return sni_pb.DetectMemoryMappingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_DetectMemoryMappingResponse(arg) {
  if (!(arg instanceof sni_pb.DetectMemoryMappingResponse)) {
    throw new Error('Expected argument of type DetectMemoryMappingResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_DetectMemoryMappingResponse(buffer_arg) {
  return sni_pb.DetectMemoryMappingResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_DevicesRequest(arg) {
  if (!(arg instanceof sni_pb.DevicesRequest)) {
    throw new Error('Expected argument of type DevicesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_DevicesRequest(buffer_arg) {
  return sni_pb.DevicesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_DevicesResponse(arg) {
  if (!(arg instanceof sni_pb.DevicesResponse)) {
    throw new Error('Expected argument of type DevicesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_DevicesResponse(buffer_arg) {
  return sni_pb.DevicesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_MultiReadMemoryRequest(arg) {
  if (!(arg instanceof sni_pb.MultiReadMemoryRequest)) {
    throw new Error('Expected argument of type MultiReadMemoryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_MultiReadMemoryRequest(buffer_arg) {
  return sni_pb.MultiReadMemoryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_MultiReadMemoryResponse(arg) {
  if (!(arg instanceof sni_pb.MultiReadMemoryResponse)) {
    throw new Error('Expected argument of type MultiReadMemoryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_MultiReadMemoryResponse(buffer_arg) {
  return sni_pb.MultiReadMemoryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_MultiWriteMemoryRequest(arg) {
  if (!(arg instanceof sni_pb.MultiWriteMemoryRequest)) {
    throw new Error('Expected argument of type MultiWriteMemoryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_MultiWriteMemoryRequest(buffer_arg) {
  return sni_pb.MultiWriteMemoryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_MultiWriteMemoryResponse(arg) {
  if (!(arg instanceof sni_pb.MultiWriteMemoryResponse)) {
    throw new Error('Expected argument of type MultiWriteMemoryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_MultiWriteMemoryResponse(buffer_arg) {
  return sni_pb.MultiWriteMemoryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_PauseEmulationRequest(arg) {
  if (!(arg instanceof sni_pb.PauseEmulationRequest)) {
    throw new Error('Expected argument of type PauseEmulationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_PauseEmulationRequest(buffer_arg) {
  return sni_pb.PauseEmulationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_PauseEmulationResponse(arg) {
  if (!(arg instanceof sni_pb.PauseEmulationResponse)) {
    throw new Error('Expected argument of type PauseEmulationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_PauseEmulationResponse(buffer_arg) {
  return sni_pb.PauseEmulationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_PauseToggleEmulationRequest(arg) {
  if (!(arg instanceof sni_pb.PauseToggleEmulationRequest)) {
    throw new Error('Expected argument of type PauseToggleEmulationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_PauseToggleEmulationRequest(buffer_arg) {
  return sni_pb.PauseToggleEmulationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_PauseToggleEmulationResponse(arg) {
  if (!(arg instanceof sni_pb.PauseToggleEmulationResponse)) {
    throw new Error('Expected argument of type PauseToggleEmulationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_PauseToggleEmulationResponse(buffer_arg) {
  return sni_pb.PauseToggleEmulationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_ResetSystemRequest(arg) {
  if (!(arg instanceof sni_pb.ResetSystemRequest)) {
    throw new Error('Expected argument of type ResetSystemRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_ResetSystemRequest(buffer_arg) {
  return sni_pb.ResetSystemRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_ResetSystemResponse(arg) {
  if (!(arg instanceof sni_pb.ResetSystemResponse)) {
    throw new Error('Expected argument of type ResetSystemResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_ResetSystemResponse(buffer_arg) {
  return sni_pb.ResetSystemResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_SingleReadMemoryRequest(arg) {
  if (!(arg instanceof sni_pb.SingleReadMemoryRequest)) {
    throw new Error('Expected argument of type SingleReadMemoryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_SingleReadMemoryRequest(buffer_arg) {
  return sni_pb.SingleReadMemoryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_SingleReadMemoryResponse(arg) {
  if (!(arg instanceof sni_pb.SingleReadMemoryResponse)) {
    throw new Error('Expected argument of type SingleReadMemoryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_SingleReadMemoryResponse(buffer_arg) {
  return sni_pb.SingleReadMemoryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_SingleWriteMemoryRequest(arg) {
  if (!(arg instanceof sni_pb.SingleWriteMemoryRequest)) {
    throw new Error('Expected argument of type SingleWriteMemoryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_SingleWriteMemoryRequest(buffer_arg) {
  return sni_pb.SingleWriteMemoryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_SingleWriteMemoryResponse(arg) {
  if (!(arg instanceof sni_pb.SingleWriteMemoryResponse)) {
    throw new Error('Expected argument of type SingleWriteMemoryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_SingleWriteMemoryResponse(buffer_arg) {
  return sni_pb.SingleWriteMemoryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var DevicesService = exports.DevicesService = {
  // detect and list devices currently connected to the system:
listDevices: {
    path: '/Devices/ListDevices',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.DevicesRequest,
    responseType: sni_pb.DevicesResponse,
    requestSerialize: serialize_DevicesRequest,
    requestDeserialize: deserialize_DevicesRequest,
    responseSerialize: serialize_DevicesResponse,
    responseDeserialize: deserialize_DevicesResponse,
  },
};

exports.DevicesClient = grpc.makeGenericClientConstructor(DevicesService);
var DeviceControlService = exports.DeviceControlService = {
  // only available if DeviceCapability ResetSystem is present
resetSystem: {
    path: '/DeviceControl/ResetSystem',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.ResetSystemRequest,
    responseType: sni_pb.ResetSystemResponse,
    requestSerialize: serialize_ResetSystemRequest,
    requestDeserialize: deserialize_ResetSystemRequest,
    responseSerialize: serialize_ResetSystemResponse,
    responseDeserialize: deserialize_ResetSystemResponse,
  },
  // only available if DeviceCapability PauseUnpauseEmulation is present
pauseUnpauseEmulation: {
    path: '/DeviceControl/PauseUnpauseEmulation',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.PauseEmulationRequest,
    responseType: sni_pb.PauseEmulationResponse,
    requestSerialize: serialize_PauseEmulationRequest,
    requestDeserialize: deserialize_PauseEmulationRequest,
    responseSerialize: serialize_PauseEmulationResponse,
    responseDeserialize: deserialize_PauseEmulationResponse,
  },
  // only available if DeviceCapability PauseToggleEmulation is present
pauseToggleEmulation: {
    path: '/DeviceControl/PauseToggleEmulation',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.PauseToggleEmulationRequest,
    responseType: sni_pb.PauseToggleEmulationResponse,
    requestSerialize: serialize_PauseToggleEmulationRequest,
    requestDeserialize: deserialize_PauseToggleEmulationRequest,
    responseSerialize: serialize_PauseToggleEmulationResponse,
    responseDeserialize: deserialize_PauseToggleEmulationResponse,
  },
};

exports.DeviceControlClient = grpc.makeGenericClientConstructor(DeviceControlService);
var DeviceMemoryService = exports.DeviceMemoryService = {
  // detect the current memory mapping for the given device by reading $00:FFB0 header:
mappingDetect: {
    path: '/DeviceMemory/MappingDetect',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.DetectMemoryMappingRequest,
    responseType: sni_pb.DetectMemoryMappingResponse,
    requestSerialize: serialize_DetectMemoryMappingRequest,
    requestDeserialize: deserialize_DetectMemoryMappingRequest,
    responseSerialize: serialize_DetectMemoryMappingResponse,
    responseDeserialize: deserialize_DetectMemoryMappingResponse,
  },
  // read a single memory segment with a given size from the given device:
singleRead: {
    path: '/DeviceMemory/SingleRead',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.SingleReadMemoryRequest,
    responseType: sni_pb.SingleReadMemoryResponse,
    requestSerialize: serialize_SingleReadMemoryRequest,
    requestDeserialize: deserialize_SingleReadMemoryRequest,
    responseSerialize: serialize_SingleReadMemoryResponse,
    responseDeserialize: deserialize_SingleReadMemoryResponse,
  },
  // write a single memory segment with given data to the given device:
singleWrite: {
    path: '/DeviceMemory/SingleWrite',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.SingleWriteMemoryRequest,
    responseType: sni_pb.SingleWriteMemoryResponse,
    requestSerialize: serialize_SingleWriteMemoryRequest,
    requestDeserialize: deserialize_SingleWriteMemoryRequest,
    responseSerialize: serialize_SingleWriteMemoryResponse,
    responseDeserialize: deserialize_SingleWriteMemoryResponse,
  },
  // read multiple memory segments with given sizes from the given device:
multiRead: {
    path: '/DeviceMemory/MultiRead',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.MultiReadMemoryRequest,
    responseType: sni_pb.MultiReadMemoryResponse,
    requestSerialize: serialize_MultiReadMemoryRequest,
    requestDeserialize: deserialize_MultiReadMemoryRequest,
    responseSerialize: serialize_MultiReadMemoryResponse,
    responseDeserialize: deserialize_MultiReadMemoryResponse,
  },
  // write multiple memory segments with given data to the given device:
multiWrite: {
    path: '/DeviceMemory/MultiWrite',
    requestStream: false,
    responseStream: false,
    requestType: sni_pb.MultiWriteMemoryRequest,
    responseType: sni_pb.MultiWriteMemoryResponse,
    requestSerialize: serialize_MultiWriteMemoryRequest,
    requestDeserialize: deserialize_MultiWriteMemoryRequest,
    responseSerialize: serialize_MultiWriteMemoryResponse,
    responseDeserialize: deserialize_MultiWriteMemoryResponse,
  },
  // stream read multiple memory segments with given sizes from the given device:
streamRead: {
    path: '/DeviceMemory/StreamRead',
    requestStream: true,
    responseStream: true,
    requestType: sni_pb.MultiReadMemoryRequest,
    responseType: sni_pb.MultiReadMemoryResponse,
    requestSerialize: serialize_MultiReadMemoryRequest,
    requestDeserialize: deserialize_MultiReadMemoryRequest,
    responseSerialize: serialize_MultiReadMemoryResponse,
    responseDeserialize: deserialize_MultiReadMemoryResponse,
  },
  // stream write multiple memory segments with given data to the given device:
streamWrite: {
    path: '/DeviceMemory/StreamWrite',
    requestStream: true,
    responseStream: true,
    requestType: sni_pb.MultiWriteMemoryRequest,
    responseType: sni_pb.MultiWriteMemoryResponse,
    requestSerialize: serialize_MultiWriteMemoryRequest,
    requestDeserialize: deserialize_MultiWriteMemoryRequest,
    responseSerialize: serialize_MultiWriteMemoryResponse,
    responseDeserialize: deserialize_MultiWriteMemoryResponse,
  },
};

exports.DeviceMemoryClient = grpc.makeGenericClientConstructor(DeviceMemoryService);
