package org.alttpo.sni

import android.app.ActivityManager
import android.content.Context
import android.os.Bundle
import android.widget.Button
import androidx.appcompat.app.AppCompatActivity


class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val btnStart = findViewById<Button>(R.id.btnStart)
        val btnStop = findViewById<Button>(R.id.btnStop)

        // set initial button state based on service running status:
        if (isServiceRunning<SNIService>()) {
            btnStart.isEnabled = false
            btnStop.isEnabled = true
        } else {
            btnStart.isEnabled = true
            btnStop.isEnabled = false
        }

        btnStart.setOnClickListener {
            SNIService.startService(this)
            btnStop.isEnabled = true
            btnStart.isEnabled = false
        }
        btnStop.setOnClickListener {
            btnStop.isEnabled = false
            btnStart.isEnabled = true
            SNIService.stopService(this)
        }
    }

    @Suppress("DEPRECATION") // Deprecated for third party Services.
    inline fun <reified T> Context.isServiceRunning() =
        (getSystemService(Context.ACTIVITY_SERVICE) as ActivityManager)
            .getRunningServices(Integer.MAX_VALUE)
            .any { it.service.className == T::class.java.name }
}
