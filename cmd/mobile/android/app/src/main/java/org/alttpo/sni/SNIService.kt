package org.alttpo.sni

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.annotation.RequiresApi
import androidx.core.app.NotificationCompat
import androidx.core.content.ContextCompat
import java.text.MessageFormat

private const val TAG = "SNIService"

enum class Actions {
    START,
    STOP
}

class SNIService : Service() {
    private val CHANNEL_ID = "SNIService Channel"

    companion object {
        fun startService(context: Context) {
            val startIntent = Intent(context, SNIService::class.java)
            startIntent.action = Actions.START.name
            //startIntent.putExtra("inputExtra", message)
            Log.i(TAG, "startForegroundService(START)")
            ContextCompat.startForegroundService(context, startIntent)
        }

        fun stopService(context: Context) {
            val stopIntent = Intent(context, SNIService::class.java)
            stopIntent.action = Actions.STOP.name
            Log.i(TAG, "startForegroundService(STOP)")
            ContextCompat.startForegroundService(context, stopIntent)
        }
    }

    @RequiresApi(Build.VERSION_CODES.M)
    override fun onCreate() {
        Log.i(TAG, "onCreate")
        super.onCreate()

        createNotificationChannel()
        val notificationIntent = Intent(this, MainActivity::class.java)
        Log.i(TAG, "PendingIntent.getActivity")
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            notificationIntent,
            PendingIntent.FLAG_IMMUTABLE
        )
        Log.i(TAG, "NotificationCompat.Builder")
        val notification = NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("SNI")
            .setContentText("Running")
            .setSmallIcon(R.drawable.ic_launcher_foreground)
            .setContentIntent(pendingIntent)
            .build()

        Log.i(TAG, "startForeground(1, notification)")
        startForeground(1, notification)
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.i(TAG, "onStartCommand executed with startId: $startId")
        if (intent != null) {
            val action = intent.action
            Log.i(TAG, "using an intent with action $action")
            when (action) {
                Actions.START.name -> startService()
                Actions.STOP.name -> stopService()
                else -> Log.i(TAG, "This should never happen. No action in the received intent")
            }
        } else {
            Log.i(TAG, "with a null intent. It has been probably restarted by the system.")
        }

        // by returning this we make sure the service is restarted if the system kills the service
        return START_STICKY
    }

    private fun startService() {
        Log.i(TAG, "startService")

        // start up SNI native service:
        Log.i(TAG, "mobile.Mobile.start()")
        mobile.Mobile.start()
    }

    private fun stopService() {
        Log.i(TAG, "stopService")
        try {
            stopForeground(true)
            stopSelf()
        } catch (e: Exception) {
            Log.i(TAG, "Service stopped without being started: ${e.message}")
        }
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            Log.i(TAG, "NotificationChannel()")
            val serviceChannel = NotificationChannel(
                CHANNEL_ID,
                "SNI Service Channel",
                NotificationManager.IMPORTANCE_DEFAULT
            )

            Log.i(TAG, "getSystemService(NotificationManager)")
            val manager = getSystemService(NotificationManager::class.java)
            Log.i(TAG, "manager.createNotificationChannel()")
            manager!!.createNotificationChannel(serviceChannel)
        } else {
            Log.i(TAG, MessageFormat.format("Build.VERSION.SDK_INT = {0}", Build.VERSION.SDK_INT))
        }
    }

    override fun onBind(intent: Intent?): IBinder? {
        Log.i(TAG, "onBind")
        return null
    }

    override fun onDestroy() {
        Log.i(TAG, "onDestroy")
        Log.i(TAG, "mobile.Mobile.stop()")
        mobile.Mobile.stop()

        super.onDestroy()
    }
}