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

class SNIService : Service() {
    private val CHANNEL_ID = "SNIService Channel"

    companion object {
        fun startService(context: Context) {
            val startIntent = Intent(context, SNIService::class.java)
            //startIntent.putExtra("inputExtra", message)
            Log.i(TAG, "startForegroundService")
            ContextCompat.startForegroundService(context, startIntent)
        }
        fun stopService(context: Context) {
            val stopIntent = Intent(context, SNIService::class.java)
            Log.i(TAG, "stopService")
            context.stopService(stopIntent)
        }
    }

    @RequiresApi(Build.VERSION_CODES.M)
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.i(TAG, "onStartCommand")

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

        // start up SNI native service:
        Log.i(TAG, "mobile.Mobile.start()")
        mobile.Mobile.start()

        return START_NOT_STICKY
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
        mobile.Mobile.stop()

        super.onDestroy()
    }
}