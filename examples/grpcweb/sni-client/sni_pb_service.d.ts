// package: 
// file: sni.proto

import * as sni_pb from "./sni_pb";
import {grpc} from "@improbable-eng/grpc-web";

type DevicesListDevices = {
  readonly methodName: string;
  readonly service: typeof Devices;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.DevicesRequest;
  readonly responseType: typeof sni_pb.DevicesResponse;
};

export class Devices {
  static readonly serviceName: string;
  static readonly ListDevices: DevicesListDevices;
}

type DeviceControlResetSystem = {
  readonly methodName: string;
  readonly service: typeof DeviceControl;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.ResetSystemRequest;
  readonly responseType: typeof sni_pb.ResetSystemResponse;
};

type DeviceControlResetToMenu = {
  readonly methodName: string;
  readonly service: typeof DeviceControl;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.ResetToMenuRequest;
  readonly responseType: typeof sni_pb.ResetToMenuResponse;
};

type DeviceControlPauseUnpauseEmulation = {
  readonly methodName: string;
  readonly service: typeof DeviceControl;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.PauseEmulationRequest;
  readonly responseType: typeof sni_pb.PauseEmulationResponse;
};

type DeviceControlPauseToggleEmulation = {
  readonly methodName: string;
  readonly service: typeof DeviceControl;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.PauseToggleEmulationRequest;
  readonly responseType: typeof sni_pb.PauseToggleEmulationResponse;
};

export class DeviceControl {
  static readonly serviceName: string;
  static readonly ResetSystem: DeviceControlResetSystem;
  static readonly ResetToMenu: DeviceControlResetToMenu;
  static readonly PauseUnpauseEmulation: DeviceControlPauseUnpauseEmulation;
  static readonly PauseToggleEmulation: DeviceControlPauseToggleEmulation;
}

type DeviceMemoryMappingDetect = {
  readonly methodName: string;
  readonly service: typeof DeviceMemory;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.DetectMemoryMappingRequest;
  readonly responseType: typeof sni_pb.DetectMemoryMappingResponse;
};

type DeviceMemorySingleRead = {
  readonly methodName: string;
  readonly service: typeof DeviceMemory;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.SingleReadMemoryRequest;
  readonly responseType: typeof sni_pb.SingleReadMemoryResponse;
};

type DeviceMemorySingleWrite = {
  readonly methodName: string;
  readonly service: typeof DeviceMemory;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.SingleWriteMemoryRequest;
  readonly responseType: typeof sni_pb.SingleWriteMemoryResponse;
};

type DeviceMemoryMultiRead = {
  readonly methodName: string;
  readonly service: typeof DeviceMemory;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.MultiReadMemoryRequest;
  readonly responseType: typeof sni_pb.MultiReadMemoryResponse;
};

type DeviceMemoryMultiWrite = {
  readonly methodName: string;
  readonly service: typeof DeviceMemory;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.MultiWriteMemoryRequest;
  readonly responseType: typeof sni_pb.MultiWriteMemoryResponse;
};

type DeviceMemoryStreamRead = {
  readonly methodName: string;
  readonly service: typeof DeviceMemory;
  readonly requestStream: true;
  readonly responseStream: true;
  readonly requestType: typeof sni_pb.MultiReadMemoryRequest;
  readonly responseType: typeof sni_pb.MultiReadMemoryResponse;
};

type DeviceMemoryStreamWrite = {
  readonly methodName: string;
  readonly service: typeof DeviceMemory;
  readonly requestStream: true;
  readonly responseStream: true;
  readonly requestType: typeof sni_pb.MultiWriteMemoryRequest;
  readonly responseType: typeof sni_pb.MultiWriteMemoryResponse;
};

export class DeviceMemory {
  static readonly serviceName: string;
  static readonly MappingDetect: DeviceMemoryMappingDetect;
  static readonly SingleRead: DeviceMemorySingleRead;
  static readonly SingleWrite: DeviceMemorySingleWrite;
  static readonly MultiRead: DeviceMemoryMultiRead;
  static readonly MultiWrite: DeviceMemoryMultiWrite;
  static readonly StreamRead: DeviceMemoryStreamRead;
  static readonly StreamWrite: DeviceMemoryStreamWrite;
}

type DeviceFilesystemReadDirectory = {
  readonly methodName: string;
  readonly service: typeof DeviceFilesystem;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.ReadDirectoryRequest;
  readonly responseType: typeof sni_pb.ReadDirectoryResponse;
};

type DeviceFilesystemMakeDirectory = {
  readonly methodName: string;
  readonly service: typeof DeviceFilesystem;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.MakeDirectoryRequest;
  readonly responseType: typeof sni_pb.MakeDirectoryResponse;
};

type DeviceFilesystemRemoveFile = {
  readonly methodName: string;
  readonly service: typeof DeviceFilesystem;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.RemoveFileRequest;
  readonly responseType: typeof sni_pb.RemoveFileResponse;
};

type DeviceFilesystemRenameFile = {
  readonly methodName: string;
  readonly service: typeof DeviceFilesystem;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.RenameFileRequest;
  readonly responseType: typeof sni_pb.RenameFileResponse;
};

type DeviceFilesystemPutFile = {
  readonly methodName: string;
  readonly service: typeof DeviceFilesystem;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.PutFileRequest;
  readonly responseType: typeof sni_pb.PutFileResponse;
};

type DeviceFilesystemGetFile = {
  readonly methodName: string;
  readonly service: typeof DeviceFilesystem;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.GetFileRequest;
  readonly responseType: typeof sni_pb.GetFileResponse;
};

type DeviceFilesystemBootFile = {
  readonly methodName: string;
  readonly service: typeof DeviceFilesystem;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.BootFileRequest;
  readonly responseType: typeof sni_pb.BootFileResponse;
};

export class DeviceFilesystem {
  static readonly serviceName: string;
  static readonly ReadDirectory: DeviceFilesystemReadDirectory;
  static readonly MakeDirectory: DeviceFilesystemMakeDirectory;
  static readonly RemoveFile: DeviceFilesystemRemoveFile;
  static readonly RenameFile: DeviceFilesystemRenameFile;
  static readonly PutFile: DeviceFilesystemPutFile;
  static readonly GetFile: DeviceFilesystemGetFile;
  static readonly BootFile: DeviceFilesystemBootFile;
}

type DeviceInfoFetchFields = {
  readonly methodName: string;
  readonly service: typeof DeviceInfo;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.FieldsRequest;
  readonly responseType: typeof sni_pb.FieldsResponse;
};

export class DeviceInfo {
  static readonly serviceName: string;
  static readonly FetchFields: DeviceInfoFetchFields;
}

type DeviceNWANWACommand = {
  readonly methodName: string;
  readonly service: typeof DeviceNWA;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof sni_pb.NWACommandRequest;
  readonly responseType: typeof sni_pb.NWACommandResponse;
};

export class DeviceNWA {
  static readonly serviceName: string;
  static readonly NWACommand: DeviceNWANWACommand;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: (status?: Status) => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: (status?: Status) => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class DevicesClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  listDevices(
    requestMessage: sni_pb.DevicesRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.DevicesResponse|null) => void
  ): UnaryResponse;
  listDevices(
    requestMessage: sni_pb.DevicesRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.DevicesResponse|null) => void
  ): UnaryResponse;
}

export class DeviceControlClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  resetSystem(
    requestMessage: sni_pb.ResetSystemRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.ResetSystemResponse|null) => void
  ): UnaryResponse;
  resetSystem(
    requestMessage: sni_pb.ResetSystemRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.ResetSystemResponse|null) => void
  ): UnaryResponse;
  resetToMenu(
    requestMessage: sni_pb.ResetToMenuRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.ResetToMenuResponse|null) => void
  ): UnaryResponse;
  resetToMenu(
    requestMessage: sni_pb.ResetToMenuRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.ResetToMenuResponse|null) => void
  ): UnaryResponse;
  pauseUnpauseEmulation(
    requestMessage: sni_pb.PauseEmulationRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.PauseEmulationResponse|null) => void
  ): UnaryResponse;
  pauseUnpauseEmulation(
    requestMessage: sni_pb.PauseEmulationRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.PauseEmulationResponse|null) => void
  ): UnaryResponse;
  pauseToggleEmulation(
    requestMessage: sni_pb.PauseToggleEmulationRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.PauseToggleEmulationResponse|null) => void
  ): UnaryResponse;
  pauseToggleEmulation(
    requestMessage: sni_pb.PauseToggleEmulationRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.PauseToggleEmulationResponse|null) => void
  ): UnaryResponse;
}

export class DeviceMemoryClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  mappingDetect(
    requestMessage: sni_pb.DetectMemoryMappingRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.DetectMemoryMappingResponse|null) => void
  ): UnaryResponse;
  mappingDetect(
    requestMessage: sni_pb.DetectMemoryMappingRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.DetectMemoryMappingResponse|null) => void
  ): UnaryResponse;
  singleRead(
    requestMessage: sni_pb.SingleReadMemoryRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.SingleReadMemoryResponse|null) => void
  ): UnaryResponse;
  singleRead(
    requestMessage: sni_pb.SingleReadMemoryRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.SingleReadMemoryResponse|null) => void
  ): UnaryResponse;
  singleWrite(
    requestMessage: sni_pb.SingleWriteMemoryRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.SingleWriteMemoryResponse|null) => void
  ): UnaryResponse;
  singleWrite(
    requestMessage: sni_pb.SingleWriteMemoryRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.SingleWriteMemoryResponse|null) => void
  ): UnaryResponse;
  multiRead(
    requestMessage: sni_pb.MultiReadMemoryRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.MultiReadMemoryResponse|null) => void
  ): UnaryResponse;
  multiRead(
    requestMessage: sni_pb.MultiReadMemoryRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.MultiReadMemoryResponse|null) => void
  ): UnaryResponse;
  multiWrite(
    requestMessage: sni_pb.MultiWriteMemoryRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.MultiWriteMemoryResponse|null) => void
  ): UnaryResponse;
  multiWrite(
    requestMessage: sni_pb.MultiWriteMemoryRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.MultiWriteMemoryResponse|null) => void
  ): UnaryResponse;
  streamRead(metadata?: grpc.Metadata): BidirectionalStream<sni_pb.MultiReadMemoryRequest, sni_pb.MultiReadMemoryResponse>;
  streamWrite(metadata?: grpc.Metadata): BidirectionalStream<sni_pb.MultiWriteMemoryRequest, sni_pb.MultiWriteMemoryResponse>;
}

export class DeviceFilesystemClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  readDirectory(
    requestMessage: sni_pb.ReadDirectoryRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.ReadDirectoryResponse|null) => void
  ): UnaryResponse;
  readDirectory(
    requestMessage: sni_pb.ReadDirectoryRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.ReadDirectoryResponse|null) => void
  ): UnaryResponse;
  makeDirectory(
    requestMessage: sni_pb.MakeDirectoryRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.MakeDirectoryResponse|null) => void
  ): UnaryResponse;
  makeDirectory(
    requestMessage: sni_pb.MakeDirectoryRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.MakeDirectoryResponse|null) => void
  ): UnaryResponse;
  removeFile(
    requestMessage: sni_pb.RemoveFileRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.RemoveFileResponse|null) => void
  ): UnaryResponse;
  removeFile(
    requestMessage: sni_pb.RemoveFileRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.RemoveFileResponse|null) => void
  ): UnaryResponse;
  renameFile(
    requestMessage: sni_pb.RenameFileRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.RenameFileResponse|null) => void
  ): UnaryResponse;
  renameFile(
    requestMessage: sni_pb.RenameFileRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.RenameFileResponse|null) => void
  ): UnaryResponse;
  putFile(
    requestMessage: sni_pb.PutFileRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.PutFileResponse|null) => void
  ): UnaryResponse;
  putFile(
    requestMessage: sni_pb.PutFileRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.PutFileResponse|null) => void
  ): UnaryResponse;
  getFile(
    requestMessage: sni_pb.GetFileRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.GetFileResponse|null) => void
  ): UnaryResponse;
  getFile(
    requestMessage: sni_pb.GetFileRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.GetFileResponse|null) => void
  ): UnaryResponse;
  bootFile(
    requestMessage: sni_pb.BootFileRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.BootFileResponse|null) => void
  ): UnaryResponse;
  bootFile(
    requestMessage: sni_pb.BootFileRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.BootFileResponse|null) => void
  ): UnaryResponse;
}

export class DeviceInfoClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  fetchFields(
    requestMessage: sni_pb.FieldsRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.FieldsResponse|null) => void
  ): UnaryResponse;
  fetchFields(
    requestMessage: sni_pb.FieldsRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.FieldsResponse|null) => void
  ): UnaryResponse;
}

export class DeviceNWAClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  nWACommand(
    requestMessage: sni_pb.NWACommandRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: sni_pb.NWACommandResponse|null) => void
  ): UnaryResponse;
  nWACommand(
    requestMessage: sni_pb.NWACommandRequest,
    callback: (error: ServiceError|null, responseMessage: sni_pb.NWACommandResponse|null) => void
  ): UnaryResponse;
}

