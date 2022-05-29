package org.alttpo.sni

import android.content.Intent
import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.widget.Button

class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val btnStart = findViewById<Button>(R.id.btnStart)
        val btnStop = findViewById<Button>(R.id.btnStop)

        btnStop.isEnabled = false

        btnStart.setOnClickListener {
            startService(Intent(this, SNIService::class.java))
            btnStop.isEnabled = true
            btnStart.isEnabled = false
        }
        btnStop.setOnClickListener {
            btnStop.isEnabled = false
            btnStart.isEnabled = true
            stopService(Intent(this, SNIService::class.java))
        }
    }
}
