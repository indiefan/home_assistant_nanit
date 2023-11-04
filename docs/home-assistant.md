# Home assistant setup guide

Configuration example:

```yaml
camera:
- name: Nanit
  platform: ffmpeg
  input: rtmp://xxx.xxx.xxx.xxx:1935/local/{your_baby_uid}

sensor:
- name: "Nanit Temperature"
  platform: mqtt
  state_topic: "nanit/babies/{your_baby_uid}/temperature"
  device_class: temperature
- name: "Nanit Humidity"
  platform: mqtt
  state_topic: "nanit/babies/{your_baby_uid}/humidity"
  device_class: humidity
```

## See also

- [Setup with NVR/Zoneminder](https://community.home-assistant.io/t/nanit-showing-in-ha-via-nvr-zoneminder/251641) by @jaburges