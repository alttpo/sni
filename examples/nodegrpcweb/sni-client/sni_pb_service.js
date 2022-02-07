// package: 
// file: sni.proto

var sni_pb = require("./sni_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var Devices = (function () {
  function Devices() {}
  Devices.serviceName = "Devices";
  return Devices;
}());

Devices.ListDevices = {
  methodName: "ListDevices",
  service: Devices,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.DevicesRequest,
  responseType: sni_pb.DevicesResponse
};

exports.Devices = Devices;

function DevicesClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

DevicesClient.prototype.listDevices = function listDevices(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(Devices.ListDevices, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

exports.DevicesClient = DevicesClient;

var DeviceControl = (function () {
  function DeviceControl() {}
  DeviceControl.serviceName = "DeviceControl";
  return DeviceControl;
}());

DeviceControl.ResetSystem = {
  methodName: "ResetSystem",
  service: DeviceControl,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.ResetSystemRequest,
  responseType: sni_pb.ResetSystemResponse
};

DeviceControl.ResetToMenu = {
  methodName: "ResetToMenu",
  service: DeviceControl,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.ResetToMenuRequest,
  responseType: sni_pb.ResetToMenuResponse
};

DeviceControl.PauseUnpauseEmulation = {
  methodName: "PauseUnpauseEmulation",
  service: DeviceControl,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.PauseEmulationRequest,
  responseType: sni_pb.PauseEmulationResponse
};

DeviceControl.PauseToggleEmulation = {
  methodName: "PauseToggleEmulation",
  service: DeviceControl,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.PauseToggleEmulationRequest,
  responseType: sni_pb.PauseToggleEmulationResponse
};

exports.DeviceControl = DeviceControl;

function DeviceControlClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

DeviceControlClient.prototype.resetSystem = function resetSystem(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceControl.ResetSystem, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceControlClient.prototype.resetToMenu = function resetToMenu(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceControl.ResetToMenu, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceControlClient.prototype.pauseUnpauseEmulation = function pauseUnpauseEmulation(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceControl.PauseUnpauseEmulation, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceControlClient.prototype.pauseToggleEmulation = function pauseToggleEmulation(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceControl.PauseToggleEmulation, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

exports.DeviceControlClient = DeviceControlClient;

var DeviceMemory = (function () {
  function DeviceMemory() {}
  DeviceMemory.serviceName = "DeviceMemory";
  return DeviceMemory;
}());

DeviceMemory.MappingDetect = {
  methodName: "MappingDetect",
  service: DeviceMemory,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.DetectMemoryMappingRequest,
  responseType: sni_pb.DetectMemoryMappingResponse
};

DeviceMemory.SingleRead = {
  methodName: "SingleRead",
  service: DeviceMemory,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.SingleReadMemoryRequest,
  responseType: sni_pb.SingleReadMemoryResponse
};

DeviceMemory.SingleWrite = {
  methodName: "SingleWrite",
  service: DeviceMemory,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.SingleWriteMemoryRequest,
  responseType: sni_pb.SingleWriteMemoryResponse
};

DeviceMemory.MultiRead = {
  methodName: "MultiRead",
  service: DeviceMemory,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.MultiReadMemoryRequest,
  responseType: sni_pb.MultiReadMemoryResponse
};

DeviceMemory.MultiWrite = {
  methodName: "MultiWrite",
  service: DeviceMemory,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.MultiWriteMemoryRequest,
  responseType: sni_pb.MultiWriteMemoryResponse
};

DeviceMemory.StreamRead = {
  methodName: "StreamRead",
  service: DeviceMemory,
  requestStream: true,
  responseStream: true,
  requestType: sni_pb.MultiReadMemoryRequest,
  responseType: sni_pb.MultiReadMemoryResponse
};

DeviceMemory.StreamWrite = {
  methodName: "StreamWrite",
  service: DeviceMemory,
  requestStream: true,
  responseStream: true,
  requestType: sni_pb.MultiWriteMemoryRequest,
  responseType: sni_pb.MultiWriteMemoryResponse
};

exports.DeviceMemory = DeviceMemory;

function DeviceMemoryClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

DeviceMemoryClient.prototype.mappingDetect = function mappingDetect(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceMemory.MappingDetect, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceMemoryClient.prototype.singleRead = function singleRead(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceMemory.SingleRead, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceMemoryClient.prototype.singleWrite = function singleWrite(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceMemory.SingleWrite, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceMemoryClient.prototype.multiRead = function multiRead(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceMemory.MultiRead, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceMemoryClient.prototype.multiWrite = function multiWrite(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceMemory.MultiWrite, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceMemoryClient.prototype.streamRead = function streamRead(metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.client(DeviceMemory.StreamRead, {
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport
  });
  client.onEnd(function (status, statusMessage, trailers) {
    listeners.status.forEach(function (handler) {
      handler({ code: status, details: statusMessage, metadata: trailers });
    });
    listeners.end.forEach(function (handler) {
      handler({ code: status, details: statusMessage, metadata: trailers });
    });
    listeners = null;
  });
  client.onMessage(function (message) {
    listeners.data.forEach(function (handler) {
      handler(message);
    })
  });
  client.start(metadata);
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    write: function (requestMessage) {
      client.send(requestMessage);
      return this;
    },
    end: function () {
      client.finishSend();
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

DeviceMemoryClient.prototype.streamWrite = function streamWrite(metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.client(DeviceMemory.StreamWrite, {
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport
  });
  client.onEnd(function (status, statusMessage, trailers) {
    listeners.status.forEach(function (handler) {
      handler({ code: status, details: statusMessage, metadata: trailers });
    });
    listeners.end.forEach(function (handler) {
      handler({ code: status, details: statusMessage, metadata: trailers });
    });
    listeners = null;
  });
  client.onMessage(function (message) {
    listeners.data.forEach(function (handler) {
      handler(message);
    })
  });
  client.start(metadata);
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    write: function (requestMessage) {
      client.send(requestMessage);
      return this;
    },
    end: function () {
      client.finishSend();
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

exports.DeviceMemoryClient = DeviceMemoryClient;

var DeviceFilesystem = (function () {
  function DeviceFilesystem() {}
  DeviceFilesystem.serviceName = "DeviceFilesystem";
  return DeviceFilesystem;
}());

DeviceFilesystem.ReadDirectory = {
  methodName: "ReadDirectory",
  service: DeviceFilesystem,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.ReadDirectoryRequest,
  responseType: sni_pb.ReadDirectoryResponse
};

DeviceFilesystem.MakeDirectory = {
  methodName: "MakeDirectory",
  service: DeviceFilesystem,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.MakeDirectoryRequest,
  responseType: sni_pb.MakeDirectoryResponse
};

DeviceFilesystem.RemoveFile = {
  methodName: "RemoveFile",
  service: DeviceFilesystem,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.RemoveFileRequest,
  responseType: sni_pb.RemoveFileResponse
};

DeviceFilesystem.RenameFile = {
  methodName: "RenameFile",
  service: DeviceFilesystem,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.RenameFileRequest,
  responseType: sni_pb.RenameFileResponse
};

DeviceFilesystem.PutFile = {
  methodName: "PutFile",
  service: DeviceFilesystem,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.PutFileRequest,
  responseType: sni_pb.PutFileResponse
};

DeviceFilesystem.GetFile = {
  methodName: "GetFile",
  service: DeviceFilesystem,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.GetFileRequest,
  responseType: sni_pb.GetFileResponse
};

DeviceFilesystem.BootFile = {
  methodName: "BootFile",
  service: DeviceFilesystem,
  requestStream: false,
  responseStream: false,
  requestType: sni_pb.BootFileRequest,
  responseType: sni_pb.BootFileResponse
};

exports.DeviceFilesystem = DeviceFilesystem;

function DeviceFilesystemClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

DeviceFilesystemClient.prototype.readDirectory = function readDirectory(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceFilesystem.ReadDirectory, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceFilesystemClient.prototype.makeDirectory = function makeDirectory(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceFilesystem.MakeDirectory, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceFilesystemClient.prototype.removeFile = function removeFile(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceFilesystem.RemoveFile, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceFilesystemClient.prototype.renameFile = function renameFile(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceFilesystem.RenameFile, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceFilesystemClient.prototype.putFile = function putFile(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceFilesystem.PutFile, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceFilesystemClient.prototype.getFile = function getFile(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceFilesystem.GetFile, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

DeviceFilesystemClient.prototype.bootFile = function bootFile(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(DeviceFilesystem.BootFile, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

exports.DeviceFilesystemClient = DeviceFilesystemClient;

