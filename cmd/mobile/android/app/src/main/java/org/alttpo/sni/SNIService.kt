package org.alttpo.sni

import android.app.Service
import android.content.Intent
import android.os.IBinder
import android.util.Log

private const val TAG = "SNIService"

class SNIService : Service() {
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.i(TAG, "onStartCommand")
        mobile.Mobile.start()
        return super.onStartCommand(intent, flags, startId)
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