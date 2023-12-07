import Devices from './devices'
import styles from './page.module.css'

export default function Home() {
  return (
    <main className={styles.main}>
      <h1>SNI gRPC Web</h1>
      <div>
        <Devices />
      </div>
    </main>
  )
}
