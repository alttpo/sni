// package: 
// file: sni.proto

import * as jspb from "google-protobuf";

export class DevicesRequest extends jspb.Message {
  clearKindsList(): void;
  getKindsList(): Array<string>;
  setKindsList(value: Array<string>): void;
  addKinds(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DevicesRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DevicesRequest): DevicesRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DevicesRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DevicesRequest;
  static deserializeBinaryFromReader(message: DevicesRequest, reader: jspb.BinaryReader): DevicesRequest;
}

export namespace DevicesRequest {
  export type AsObject = {
    kindsList: Array<string>,
  }
}

export class DevicesResponse extends jspb.Message {
  clearDevicesList(): void;
  getDevicesList(): Array<DevicesResponse.Device>;
  setDevicesList(value: Array<DevicesResponse.Device>): void;
  addDevices(value?: DevicesResponse.Device, index?: number): DevicesResponse.Device;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DevicesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DevicesResponse): DevicesResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DevicesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DevicesResponse;
  static deserializeBinaryFromReader(message: DevicesResponse, reader: jspb.BinaryReader): DevicesResponse;
}

export namespace DevicesResponse {
  export type AsObject = {
    devicesList: Array<DevicesResponse.Device.AsObject>,
  }

  export class Device extends jspb.Message {
    getUri(): string;
    setUri(value: string): void;

    getDisplayname(): string;
    setDisplayname(value: string): void;

    getKind(): string;
    setKind(value: string): void;

    clearCapabilitiesList(): void;
    getCapabilitiesList(): Array<DeviceCapabilityMap[keyof DeviceCapabilityMap]>;
    setCapabilitiesList(value: Array<DeviceCapabilityMap[keyof DeviceCapabilityMap]>): void;
    addCapabilities(value: DeviceCapabilityMap[keyof DeviceCapabilityMap], index?: number): DeviceCapabilityMap[keyof DeviceCapabilityMap];

    getDefaultaddressspace(): AddressSpaceMap[keyof AddressSpaceMap];
    setDefaultaddressspace(value: AddressSpaceMap[keyof AddressSpaceMap]): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Device.AsObject;
    static toObject(includeInstance: boolean, msg: Device): Device.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Device, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Device;
    static deserializeBinaryFromReader(message: Device, reader: jspb.BinaryReader): Device;
  }

  export namespace Device {
    export type AsObject = {
      uri: string,
      displayname: string,
      kind: string,
      capabilitiesList: Array<DeviceCapabilityMap[keyof DeviceCapabilityMap]>,
      defaultaddressspace: AddressSpaceMap[keyof AddressSpaceMap],
    }
  }
}

export class ResetSystemRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResetSystemRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ResetSystemRequest): ResetSystemRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ResetSystemRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResetSystemRequest;
  static deserializeBinaryFromReader(message: ResetSystemRequest, reader: jspb.BinaryReader): ResetSystemRequest;
}

export namespace ResetSystemRequest {
  export type AsObject = {
    uri: string,
  }
}

export class ResetSystemResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResetSystemResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ResetSystemResponse): ResetSystemResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ResetSystemResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResetSystemResponse;
  static deserializeBinaryFromReader(message: ResetSystemResponse, reader: jspb.BinaryReader): ResetSystemResponse;
}

export namespace ResetSystemResponse {
  export type AsObject = {
    uri: string,
  }
}

export class ResetToMenuRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResetToMenuRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ResetToMenuRequest): ResetToMenuRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ResetToMenuRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResetToMenuRequest;
  static deserializeBinaryFromReader(message: ResetToMenuRequest, reader: jspb.BinaryReader): ResetToMenuRequest;
}

export namespace ResetToMenuRequest {
  export type AsObject = {
    uri: string,
  }
}

export class ResetToMenuResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResetToMenuResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ResetToMenuResponse): ResetToMenuResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ResetToMenuResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResetToMenuResponse;
  static deserializeBinaryFromReader(message: ResetToMenuResponse, reader: jspb.BinaryReader): ResetToMenuResponse;
}

export namespace ResetToMenuResponse {
  export type AsObject = {
    uri: string,
  }
}

export class PauseEmulationRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPaused(): boolean;
  setPaused(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PauseEmulationRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PauseEmulationRequest): PauseEmulationRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PauseEmulationRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PauseEmulationRequest;
  static deserializeBinaryFromReader(message: PauseEmulationRequest, reader: jspb.BinaryReader): PauseEmulationRequest;
}

export namespace PauseEmulationRequest {
  export type AsObject = {
    uri: string,
    paused: boolean,
  }
}

export class PauseEmulationResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPaused(): boolean;
  setPaused(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PauseEmulationResponse.AsObject;
  static toObject(includeInstance: boolean, msg: PauseEmulationResponse): PauseEmulationResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PauseEmulationResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PauseEmulationResponse;
  static deserializeBinaryFromReader(message: PauseEmulationResponse, reader: jspb.BinaryReader): PauseEmulationResponse;
}

export namespace PauseEmulationResponse {
  export type AsObject = {
    uri: string,
    paused: boolean,
  }
}

export class PauseToggleEmulationRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PauseToggleEmulationRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PauseToggleEmulationRequest): PauseToggleEmulationRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PauseToggleEmulationRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PauseToggleEmulationRequest;
  static deserializeBinaryFromReader(message: PauseToggleEmulationRequest, reader: jspb.BinaryReader): PauseToggleEmulationRequest;
}

export namespace PauseToggleEmulationRequest {
  export type AsObject = {
    uri: string,
  }
}

export class PauseToggleEmulationResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PauseToggleEmulationResponse.AsObject;
  static toObject(includeInstance: boolean, msg: PauseToggleEmulationResponse): PauseToggleEmulationResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PauseToggleEmulationResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PauseToggleEmulationResponse;
  static deserializeBinaryFromReader(message: PauseToggleEmulationResponse, reader: jspb.BinaryReader): PauseToggleEmulationResponse;
}

export namespace PauseToggleEmulationResponse {
  export type AsObject = {
    uri: string,
  }
}

export class DetectMemoryMappingRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  hasFallbackmemorymapping(): boolean;
  clearFallbackmemorymapping(): void;
  getFallbackmemorymapping(): MemoryMappingMap[keyof MemoryMappingMap];
  setFallbackmemorymapping(value: MemoryMappingMap[keyof MemoryMappingMap]): void;

  hasRomheader00ffb0(): boolean;
  clearRomheader00ffb0(): void;
  getRomheader00ffb0(): Uint8Array | string;
  getRomheader00ffb0_asU8(): Uint8Array;
  getRomheader00ffb0_asB64(): string;
  setRomheader00ffb0(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DetectMemoryMappingRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DetectMemoryMappingRequest): DetectMemoryMappingRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DetectMemoryMappingRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DetectMemoryMappingRequest;
  static deserializeBinaryFromReader(message: DetectMemoryMappingRequest, reader: jspb.BinaryReader): DetectMemoryMappingRequest;
}

export namespace DetectMemoryMappingRequest {
  export type AsObject = {
    uri: string,
    fallbackmemorymapping: MemoryMappingMap[keyof MemoryMappingMap],
    romheader00ffb0: Uint8Array | string,
  }
}

export class DetectMemoryMappingResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getMemorymapping(): MemoryMappingMap[keyof MemoryMappingMap];
  setMemorymapping(value: MemoryMappingMap[keyof MemoryMappingMap]): void;

  getConfidence(): boolean;
  setConfidence(value: boolean): void;

  getRomheader00ffb0(): Uint8Array | string;
  getRomheader00ffb0_asU8(): Uint8Array;
  getRomheader00ffb0_asB64(): string;
  setRomheader00ffb0(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DetectMemoryMappingResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DetectMemoryMappingResponse): DetectMemoryMappingResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DetectMemoryMappingResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DetectMemoryMappingResponse;
  static deserializeBinaryFromReader(message: DetectMemoryMappingResponse, reader: jspb.BinaryReader): DetectMemoryMappingResponse;
}

export namespace DetectMemoryMappingResponse {
  export type AsObject = {
    uri: string,
    memorymapping: MemoryMappingMap[keyof MemoryMappingMap],
    confidence: boolean,
    romheader00ffb0: Uint8Array | string,
  }
}

export class ReadMemoryRequest extends jspb.Message {
  getRequestaddress(): number;
  setRequestaddress(value: number): void;

  getRequestaddressspace(): AddressSpaceMap[keyof AddressSpaceMap];
  setRequestaddressspace(value: AddressSpaceMap[keyof AddressSpaceMap]): void;

  getRequestmemorymapping(): MemoryMappingMap[keyof MemoryMappingMap];
  setRequestmemorymapping(value: MemoryMappingMap[keyof MemoryMappingMap]): void;

  getSize(): number;
  setSize(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadMemoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ReadMemoryRequest): ReadMemoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ReadMemoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadMemoryRequest;
  static deserializeBinaryFromReader(message: ReadMemoryRequest, reader: jspb.BinaryReader): ReadMemoryRequest;
}

export namespace ReadMemoryRequest {
  export type AsObject = {
    requestaddress: number,
    requestaddressspace: AddressSpaceMap[keyof AddressSpaceMap],
    requestmemorymapping: MemoryMappingMap[keyof MemoryMappingMap],
    size: number,
  }
}

export class ReadMemoryResponse extends jspb.Message {
  getRequestaddress(): number;
  setRequestaddress(value: number): void;

  getRequestaddressspace(): AddressSpaceMap[keyof AddressSpaceMap];
  setRequestaddressspace(value: AddressSpaceMap[keyof AddressSpaceMap]): void;

  getRequestmemorymapping(): MemoryMappingMap[keyof MemoryMappingMap];
  setRequestmemorymapping(value: MemoryMappingMap[keyof MemoryMappingMap]): void;

  getDeviceaddress(): number;
  setDeviceaddress(value: number): void;

  getDeviceaddressspace(): AddressSpaceMap[keyof AddressSpaceMap];
  setDeviceaddressspace(value: AddressSpaceMap[keyof AddressSpaceMap]): void;

  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadMemoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ReadMemoryResponse): ReadMemoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ReadMemoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadMemoryResponse;
  static deserializeBinaryFromReader(message: ReadMemoryResponse, reader: jspb.BinaryReader): ReadMemoryResponse;
}

export namespace ReadMemoryResponse {
  export type AsObject = {
    requestaddress: number,
    requestaddressspace: AddressSpaceMap[keyof AddressSpaceMap],
    requestmemorymapping: MemoryMappingMap[keyof MemoryMappingMap],
    deviceaddress: number,
    deviceaddressspace: AddressSpaceMap[keyof AddressSpaceMap],
    data: Uint8Array | string,
  }
}

export class WriteMemoryRequest extends jspb.Message {
  getRequestaddress(): number;
  setRequestaddress(value: number): void;

  getRequestaddressspace(): AddressSpaceMap[keyof AddressSpaceMap];
  setRequestaddressspace(value: AddressSpaceMap[keyof AddressSpaceMap]): void;

  getRequestmemorymapping(): MemoryMappingMap[keyof MemoryMappingMap];
  setRequestmemorymapping(value: MemoryMappingMap[keyof MemoryMappingMap]): void;

  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WriteMemoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: WriteMemoryRequest): WriteMemoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WriteMemoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WriteMemoryRequest;
  static deserializeBinaryFromReader(message: WriteMemoryRequest, reader: jspb.BinaryReader): WriteMemoryRequest;
}

export namespace WriteMemoryRequest {
  export type AsObject = {
    requestaddress: number,
    requestaddressspace: AddressSpaceMap[keyof AddressSpaceMap],
    requestmemorymapping: MemoryMappingMap[keyof MemoryMappingMap],
    data: Uint8Array | string,
  }
}

export class WriteMemoryResponse extends jspb.Message {
  getRequestaddress(): number;
  setRequestaddress(value: number): void;

  getRequestaddressspace(): AddressSpaceMap[keyof AddressSpaceMap];
  setRequestaddressspace(value: AddressSpaceMap[keyof AddressSpaceMap]): void;

  getRequestmemorymapping(): MemoryMappingMap[keyof MemoryMappingMap];
  setRequestmemorymapping(value: MemoryMappingMap[keyof MemoryMappingMap]): void;

  getDeviceaddress(): number;
  setDeviceaddress(value: number): void;

  getDeviceaddressspace(): AddressSpaceMap[keyof AddressSpaceMap];
  setDeviceaddressspace(value: AddressSpaceMap[keyof AddressSpaceMap]): void;

  getSize(): number;
  setSize(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WriteMemoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: WriteMemoryResponse): WriteMemoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WriteMemoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WriteMemoryResponse;
  static deserializeBinaryFromReader(message: WriteMemoryResponse, reader: jspb.BinaryReader): WriteMemoryResponse;
}

export namespace WriteMemoryResponse {
  export type AsObject = {
    requestaddress: number,
    requestaddressspace: AddressSpaceMap[keyof AddressSpaceMap],
    requestmemorymapping: MemoryMappingMap[keyof MemoryMappingMap],
    deviceaddress: number,
    deviceaddressspace: AddressSpaceMap[keyof AddressSpaceMap],
    size: number,
  }
}

export class SingleReadMemoryRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  hasRequest(): boolean;
  clearRequest(): void;
  getRequest(): ReadMemoryRequest | undefined;
  setRequest(value?: ReadMemoryRequest): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SingleReadMemoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SingleReadMemoryRequest): SingleReadMemoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SingleReadMemoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SingleReadMemoryRequest;
  static deserializeBinaryFromReader(message: SingleReadMemoryRequest, reader: jspb.BinaryReader): SingleReadMemoryRequest;
}

export namespace SingleReadMemoryRequest {
  export type AsObject = {
    uri: string,
    request?: ReadMemoryRequest.AsObject,
  }
}

export class SingleReadMemoryResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  hasResponse(): boolean;
  clearResponse(): void;
  getResponse(): ReadMemoryResponse | undefined;
  setResponse(value?: ReadMemoryResponse): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SingleReadMemoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SingleReadMemoryResponse): SingleReadMemoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SingleReadMemoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SingleReadMemoryResponse;
  static deserializeBinaryFromReader(message: SingleReadMemoryResponse, reader: jspb.BinaryReader): SingleReadMemoryResponse;
}

export namespace SingleReadMemoryResponse {
  export type AsObject = {
    uri: string,
    response?: ReadMemoryResponse.AsObject,
  }
}

export class SingleWriteMemoryRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  hasRequest(): boolean;
  clearRequest(): void;
  getRequest(): WriteMemoryRequest | undefined;
  setRequest(value?: WriteMemoryRequest): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SingleWriteMemoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SingleWriteMemoryRequest): SingleWriteMemoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SingleWriteMemoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SingleWriteMemoryRequest;
  static deserializeBinaryFromReader(message: SingleWriteMemoryRequest, reader: jspb.BinaryReader): SingleWriteMemoryRequest;
}

export namespace SingleWriteMemoryRequest {
  export type AsObject = {
    uri: string,
    request?: WriteMemoryRequest.AsObject,
  }
}

export class SingleWriteMemoryResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  hasResponse(): boolean;
  clearResponse(): void;
  getResponse(): WriteMemoryResponse | undefined;
  setResponse(value?: WriteMemoryResponse): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SingleWriteMemoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SingleWriteMemoryResponse): SingleWriteMemoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SingleWriteMemoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SingleWriteMemoryResponse;
  static deserializeBinaryFromReader(message: SingleWriteMemoryResponse, reader: jspb.BinaryReader): SingleWriteMemoryResponse;
}

export namespace SingleWriteMemoryResponse {
  export type AsObject = {
    uri: string,
    response?: WriteMemoryResponse.AsObject,
  }
}

export class MultiReadMemoryRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  clearRequestsList(): void;
  getRequestsList(): Array<ReadMemoryRequest>;
  setRequestsList(value: Array<ReadMemoryRequest>): void;
  addRequests(value?: ReadMemoryRequest, index?: number): ReadMemoryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MultiReadMemoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: MultiReadMemoryRequest): MultiReadMemoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MultiReadMemoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MultiReadMemoryRequest;
  static deserializeBinaryFromReader(message: MultiReadMemoryRequest, reader: jspb.BinaryReader): MultiReadMemoryRequest;
}

export namespace MultiReadMemoryRequest {
  export type AsObject = {
    uri: string,
    requestsList: Array<ReadMemoryRequest.AsObject>,
  }
}

export class MultiReadMemoryResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  clearResponsesList(): void;
  getResponsesList(): Array<ReadMemoryResponse>;
  setResponsesList(value: Array<ReadMemoryResponse>): void;
  addResponses(value?: ReadMemoryResponse, index?: number): ReadMemoryResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MultiReadMemoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: MultiReadMemoryResponse): MultiReadMemoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MultiReadMemoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MultiReadMemoryResponse;
  static deserializeBinaryFromReader(message: MultiReadMemoryResponse, reader: jspb.BinaryReader): MultiReadMemoryResponse;
}

export namespace MultiReadMemoryResponse {
  export type AsObject = {
    uri: string,
    responsesList: Array<ReadMemoryResponse.AsObject>,
  }
}

export class MultiWriteMemoryRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  clearRequestsList(): void;
  getRequestsList(): Array<WriteMemoryRequest>;
  setRequestsList(value: Array<WriteMemoryRequest>): void;
  addRequests(value?: WriteMemoryRequest, index?: number): WriteMemoryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MultiWriteMemoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: MultiWriteMemoryRequest): MultiWriteMemoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MultiWriteMemoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MultiWriteMemoryRequest;
  static deserializeBinaryFromReader(message: MultiWriteMemoryRequest, reader: jspb.BinaryReader): MultiWriteMemoryRequest;
}

export namespace MultiWriteMemoryRequest {
  export type AsObject = {
    uri: string,
    requestsList: Array<WriteMemoryRequest.AsObject>,
  }
}

export class MultiWriteMemoryResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  clearResponsesList(): void;
  getResponsesList(): Array<WriteMemoryResponse>;
  setResponsesList(value: Array<WriteMemoryResponse>): void;
  addResponses(value?: WriteMemoryResponse, index?: number): WriteMemoryResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MultiWriteMemoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: MultiWriteMemoryResponse): MultiWriteMemoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MultiWriteMemoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MultiWriteMemoryResponse;
  static deserializeBinaryFromReader(message: MultiWriteMemoryResponse, reader: jspb.BinaryReader): MultiWriteMemoryResponse;
}

export namespace MultiWriteMemoryResponse {
  export type AsObject = {
    uri: string,
    responsesList: Array<WriteMemoryResponse.AsObject>,
  }
}

export class ReadDirectoryRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadDirectoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ReadDirectoryRequest): ReadDirectoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ReadDirectoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadDirectoryRequest;
  static deserializeBinaryFromReader(message: ReadDirectoryRequest, reader: jspb.BinaryReader): ReadDirectoryRequest;
}

export namespace ReadDirectoryRequest {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export class DirEntry extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  getType(): DirEntryTypeMap[keyof DirEntryTypeMap];
  setType(value: DirEntryTypeMap[keyof DirEntryTypeMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DirEntry.AsObject;
  static toObject(includeInstance: boolean, msg: DirEntry): DirEntry.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DirEntry, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DirEntry;
  static deserializeBinaryFromReader(message: DirEntry, reader: jspb.BinaryReader): DirEntry;
}

export namespace DirEntry {
  export type AsObject = {
    name: string,
    type: DirEntryTypeMap[keyof DirEntryTypeMap],
  }
}

export class ReadDirectoryResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  clearEntriesList(): void;
  getEntriesList(): Array<DirEntry>;
  setEntriesList(value: Array<DirEntry>): void;
  addEntries(value?: DirEntry, index?: number): DirEntry;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadDirectoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ReadDirectoryResponse): ReadDirectoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ReadDirectoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadDirectoryResponse;
  static deserializeBinaryFromReader(message: ReadDirectoryResponse, reader: jspb.BinaryReader): ReadDirectoryResponse;
}

export namespace ReadDirectoryResponse {
  export type AsObject = {
    uri: string,
    path: string,
    entriesList: Array<DirEntry.AsObject>,
  }
}

export class MakeDirectoryRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MakeDirectoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: MakeDirectoryRequest): MakeDirectoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MakeDirectoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MakeDirectoryRequest;
  static deserializeBinaryFromReader(message: MakeDirectoryRequest, reader: jspb.BinaryReader): MakeDirectoryRequest;
}

export namespace MakeDirectoryRequest {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export class MakeDirectoryResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MakeDirectoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: MakeDirectoryResponse): MakeDirectoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MakeDirectoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MakeDirectoryResponse;
  static deserializeBinaryFromReader(message: MakeDirectoryResponse, reader: jspb.BinaryReader): MakeDirectoryResponse;
}

export namespace MakeDirectoryResponse {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export class RemoveFileRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoveFileRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RemoveFileRequest): RemoveFileRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RemoveFileRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RemoveFileRequest;
  static deserializeBinaryFromReader(message: RemoveFileRequest, reader: jspb.BinaryReader): RemoveFileRequest;
}

export namespace RemoveFileRequest {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export class RemoveFileResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoveFileResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RemoveFileResponse): RemoveFileResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RemoveFileResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RemoveFileResponse;
  static deserializeBinaryFromReader(message: RemoveFileResponse, reader: jspb.BinaryReader): RemoveFileResponse;
}

export namespace RemoveFileResponse {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export class RenameFileRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  getNewfilename(): string;
  setNewfilename(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RenameFileRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RenameFileRequest): RenameFileRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RenameFileRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RenameFileRequest;
  static deserializeBinaryFromReader(message: RenameFileRequest, reader: jspb.BinaryReader): RenameFileRequest;
}

export namespace RenameFileRequest {
  export type AsObject = {
    uri: string,
    path: string,
    newfilename: string,
  }
}

export class RenameFileResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  getNewfilename(): string;
  setNewfilename(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RenameFileResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RenameFileResponse): RenameFileResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RenameFileResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RenameFileResponse;
  static deserializeBinaryFromReader(message: RenameFileResponse, reader: jspb.BinaryReader): RenameFileResponse;
}

export namespace RenameFileResponse {
  export type AsObject = {
    uri: string,
    path: string,
    newfilename: string,
  }
}

export class PutFileRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PutFileRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PutFileRequest): PutFileRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PutFileRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PutFileRequest;
  static deserializeBinaryFromReader(message: PutFileRequest, reader: jspb.BinaryReader): PutFileRequest;
}

export namespace PutFileRequest {
  export type AsObject = {
    uri: string,
    path: string,
    data: Uint8Array | string,
  }
}

export class PutFileResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  getSize(): number;
  setSize(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PutFileResponse.AsObject;
  static toObject(includeInstance: boolean, msg: PutFileResponse): PutFileResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PutFileResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PutFileResponse;
  static deserializeBinaryFromReader(message: PutFileResponse, reader: jspb.BinaryReader): PutFileResponse;
}

export namespace PutFileResponse {
  export type AsObject = {
    uri: string,
    path: string,
    size: number,
  }
}

export class GetFileRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetFileRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetFileRequest): GetFileRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetFileRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetFileRequest;
  static deserializeBinaryFromReader(message: GetFileRequest, reader: jspb.BinaryReader): GetFileRequest;
}

export namespace GetFileRequest {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export class GetFileResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  getSize(): number;
  setSize(value: number): void;

  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetFileResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetFileResponse): GetFileResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetFileResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetFileResponse;
  static deserializeBinaryFromReader(message: GetFileResponse, reader: jspb.BinaryReader): GetFileResponse;
}

export namespace GetFileResponse {
  export type AsObject = {
    uri: string,
    path: string,
    size: number,
    data: Uint8Array | string,
  }
}

export class BootFileRequest extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BootFileRequest.AsObject;
  static toObject(includeInstance: boolean, msg: BootFileRequest): BootFileRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BootFileRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BootFileRequest;
  static deserializeBinaryFromReader(message: BootFileRequest, reader: jspb.BinaryReader): BootFileRequest;
}

export namespace BootFileRequest {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export class BootFileResponse extends jspb.Message {
  getUri(): string;
  setUri(value: string): void;

  getPath(): string;
  setPath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BootFileResponse.AsObject;
  static toObject(includeInstance: boolean, msg: BootFileResponse): BootFileResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BootFileResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BootFileResponse;
  static deserializeBinaryFromReader(message: BootFileResponse, reader: jspb.BinaryReader): BootFileResponse;
}

export namespace BootFileResponse {
  export type AsObject = {
    uri: string,
    path: string,
  }
}

export interface AddressSpaceMap {
  FXPAKPRO: 0;
  SNESABUS: 1;
  RAW: 2;
}

export const AddressSpace: AddressSpaceMap;

export interface MemoryMappingMap {
  UNKNOWN: 0;
  HIROM: 1;
  LOROM: 2;
  EXHIROM: 3;
}

export const MemoryMapping: MemoryMappingMap;

export interface DeviceCapabilityMap {
  NONE: 0;
  READMEMORY: 1;
  WRITEMEMORY: 2;
  EXECUTEASM: 3;
  RESETSYSTEM: 4;
  PAUSEUNPAUSEEMULATION: 5;
  PAUSETOGGLEEMULATION: 6;
  RESETTOMENU: 7;
  READDIRECTORY: 10;
  MAKEDIRECTORY: 11;
  REMOVEFILE: 12;
  RENAMEFILE: 13;
  PUTFILE: 14;
  GETFILE: 15;
  BOOTFILE: 16;
}

export const DeviceCapability: DeviceCapabilityMap;

export interface DirEntryTypeMap {
  DIRECTORY: 0;
  FILE: 1;
}

export const DirEntryType: DirEntryTypeMap;

