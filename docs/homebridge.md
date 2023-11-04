# Homebridge setup guide

1. Install plugin [Homebridge Camera FFmpeg](https://github.com/Sunoo/homebridge-camera-ffmpeg#readme)
2. Configure camera in UI (or in config similar to following snippet):

```json
{
  "name": "Camera FFmpeg",
  "cameras": [
    {
      "name": "Nanit",
      "videoConfig": {
        "source": "-i rtmp://xxx.xxx.xxx.xxx:1935/local/{your_baby_uid}",
        "maxWidth": 1280,
        "maxHeight": 960,
        "maxFPS": 10,
        "maxBitrate": 3145,
        "audio": true
      }
    }
  ],
  "platform": "Camera-ffmpeg"
}
```

3. Restart Homebridge