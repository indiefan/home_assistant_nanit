# Sensors

App exposes cam sensors by publishing the updates to MQTT. See `NANIT_MQTT_*` variables in the [.env.sample](../.env.sample) file for configuration.

It will push any sensor updates to following topics:

- `nanit/babies/{baby_uid}/temperature` - temperature in degrees celsius (float)
- `nanit/babies/{baby_uid}/humidity` - humidity in percent (float)
- `nanit/babies/{baby_uid}/is_night` - flag if cam is in the night mode (bool)

You can configure these in your [HASS setup](./home-assistant.md).

In case you run into trouble and need to see what is going on, you can try using [MQTT Explorer](http://mqtt-explorer.com/).