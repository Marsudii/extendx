package com.example.extendx;

import android.app.*;
import android.content.Intent;
import android.graphics.Bitmap;
import android.media.Image;
import android.media.ImageReader;
import android.media.projection.MediaProjection;
import android.media.projection.MediaProjectionManager;
import android.os.Build;
import android.os.IBinder;
import android.util.DisplayMetrics;
import android.util.Log;
import android.view.Surface;
import android.view.WindowManager;

import androidx.annotation.Nullable;
import androidx.core.app.NotificationCompat;

import java.io.ByteArrayOutputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.nio.ByteBuffer;

public class ScreenCaptureService extends Service {

    private MediaProjection mediaProjection;
    private ImageReader imageReader;

    @Override
    public void onCreate() {
        super.onCreate();
        createNotification();
    }

    private void createNotification() {
        String CHANNEL_ID = "screen_capture_channel";
        NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID, "Screen Capture", NotificationManager.IMPORTANCE_LOW);
        NotificationManager manager = getSystemService(NotificationManager.class);
        manager.createNotificationChannel(channel);

        Notification notification = new NotificationCompat.Builder(this, CHANNEL_ID)
                .setContentTitle("ExtendX Capture Aktif")
                .setContentText("Layar sedang dibagikan")
                .setSmallIcon(android.R.drawable.ic_media_play)
                .build();

        startForeground(1, notification);
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        int resultCode = intent.getIntExtra("resultCode", Activity.RESULT_CANCELED);
        Intent data = intent.getParcelableExtra("data");

        MediaProjectionManager manager = (MediaProjectionManager) getSystemService(MEDIA_PROJECTION_SERVICE);
        mediaProjection = manager.getMediaProjection(resultCode, data);

        startCapture();

        return START_STICKY;
    }

    private void startCapture() {
        DisplayMetrics metrics = new DisplayMetrics();
        WindowManager wm = (WindowManager) getSystemService(WINDOW_SERVICE);
        wm.getDefaultDisplay().getMetrics(metrics);

        int width = metrics.widthPixels;
        int height = metrics.heightPixels;

        imageReader = ImageReader.newInstance(width, height, 0x1, 2);
        Surface surface = imageReader.getSurface();

        mediaProjection.createVirtualDisplay("ScreenCapture", width, height, metrics.densityDpi,
                0, surface, null, null);

        imageReader.setOnImageAvailableListener(reader -> {
            Image image = reader.acquireLatestImage();
            if (image != null) {
                Image.Plane[] planes = image.getPlanes();
                if (planes.length > 0) {
                    ByteBuffer buffer = planes[0].getBuffer();
                    int pixelStride = planes[0].getPixelStride();
                    int rowStride = planes[0].getRowStride();
                    int rowPadding = rowStride - pixelStride * width;

                    Bitmap bitmap = Bitmap.createBitmap(width + rowPadding / pixelStride,
                            height, Bitmap.Config.ARGB_8888);
                    bitmap.copyPixelsFromBuffer(buffer);
                    sendBitmapOverTCP(bitmap);
                }
                image.close();
            }
        }, null);
    }

    private void sendBitmapOverTCP(Bitmap bitmap) {
        new Thread(() -> {
            try {
                Socket socket = new Socket("10.10.10.5", 8088); // Ganti dengan IP client
                OutputStream out = socket.getOutputStream();

                ByteArrayOutputStream baos = new ByteArrayOutputStream();
                bitmap.compress(Bitmap.CompressFormat.JPEG, 70, baos);
                byte[] jpegData = baos.toByteArray();

                ByteBuffer sizeBuffer = ByteBuffer.allocate(4).putInt(jpegData.length);
                out.write(sizeBuffer.array());
                out.write(jpegData);
                out.flush();
                socket.close();
            } catch (Exception e) {
                Log.e("TCP", "Send Error: " + e.getMessage());
            }
        }).start();
    }

    @Nullable
    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}