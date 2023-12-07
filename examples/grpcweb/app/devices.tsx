'use client'

import * as SNI from '@/lib/sni'
import * as SNIClient from '@/lib/sni.client'
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'
import { DevicesResponse_Device } from '@/lib/sni'
import { useState } from 'react'
import styles from './devices.module.css'

let transport = new GrpcWebFetchTransport({
  baseUrl: 'http://localhost:8190'
})

const listDevices = async () => {
  const DevicesClient = new SNIClient.DevicesClient(transport)
  const req = SNI.DevicesRequest.create()
  const devices = await DevicesClient.listDevices(req)
  return devices.response.devices
}

const DeviceButton = ({ onUpdate }: { onUpdate: (devices: any) => void }) => {
  return (
    <button
      className={styles.btn}
      onClick={async (evt) => {
      evt.preventDefault()
      const devices = await listDevices()
      onUpdate(devices)
    }}>
      List Devices
    </button>
  )
}

const Device = ({ displayName = '', ...props }: DevicesResponse_Device) => {
  const [expanded, setExpanded] = useState(false)
  return (
    <li>
      <div className={styles.device_name}>
        <span className={styles.device_label}>{displayName}</span>
      </div>
      <button
        className={`${styles.btn} ${styles.secondary}`}
        onClick={(evt) => {
          evt.preventDefault()
          setExpanded(!expanded)
        }}
      >
        {expanded ? 'Hide full data' : 'View full data'}
      </button>
      {expanded && (
        <div className={styles.json}>
          <pre>
            {JSON.stringify(props, null, 2)}
          </pre>
        </div>
      )}
    </li>
  )
}

const DeviceList = ({ devices }: { devices: DevicesResponse_Device[]|null }) => {
  if (!devices) {
    return null
  }

  if (devices.length === 0) {
    return (
      <div style={{ textAlign: 'center' }}>
        No devices found
      </div>
    )
  }
  return (
    <ol className={styles.list}>
      {devices.map((device) => (
        <Device key={device.uri} {...device} />
      ))}
    </ol>
  )
}

const Devices = () => {
  const [devices, setDevices] = useState(null)

  return (
    <div className={styles.container}>
      <div className={styles.actionContainer}>
        <DeviceButton onUpdate={setDevices} />
      </div>
      {devices && (
        <div>
          <DeviceList devices={devices} />
        </div>
      )}
    </div>
  )
}

export default Devices
