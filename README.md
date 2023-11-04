# Background

This is a fork of a no-longer-maintained project (https://gitlab.com/adam.stanek/nanit) with added support for Nanit's (now required) 2FA authentication.


Original readme below:

# Nanit Stream Proxy

This is sleepless night induced pet project to restream Nanit Baby Monitor live stream for local viewing.

## Features

- Restreaming of live feed to local RTMP server
- Retrieving sensors data from cam (temperature and humidity) and publishing them over MQTT
- Graceful authentication session handling
- Works as a companion for your Home-assistant / Homebridge setup (see [guides](#setup-guides) below)

## TL;DR

```bash
# Note: use your local IP, reachable from Cam (not 127.0.0.1)

docker run --rm \
  -e NANIT_EMAIL=your@email.tld \
  -e NANIT_PASSWORD=XXXXXXXXXXXXX \
  -e NANIT_RTMP_ADDR=xxx.xxx.xxx.xxx:1935 \
  -p 1935:1935 \
  registry.gitlab.com/adam.stanek/nanit:v0-7
```

Open `rtmp://127.0.0.1:1935/local/[your_baby_uid]` in VLC. You will find your baby UID in the log of running application.
### Setup guides

- [Home assistant](./docs/home-assistant.md)
- [Homebridge](./docs/homebridge.md)
- [Sensors](./docs/sensors.md)
- [Docker compose](./docs/docker-compose.md)

### Further usage

Application is ready to be used in Docker. You can use environment variables for configuration. For more info see [.env.sample](.env.sample).
## Why?

- I wanted to learn something new on paternity leave (first project in Go!)
- Nanit iOS application is nice, but I was really disappointed that it cannot properly stream to TV through AirPlay. As anxious parents of our first child we wanted to have it playing in the background on TV when we are in the kitchen, etc. When AirPlaying it from the phone it was really hard to see the little one in portrait mode + the sound was crazy quiet. This helps us around the issue and we don't have to drain our phone batteries.

## How to develop

```bash
go run cmd/nanit/*.go

# On proto file change
protoc --go_out . --go_opt=paths=source_relative pkg/client/websocket.proto

# Run tests
go test ./pkg/...
```

For some insights see [Developer notes](docs/developer-notes.md).

## Disclaimer

I made this program solely for learning purposes. Please use it at your own risk and always follow any terms and conditions which might be applied when communicating to Nanit servers.

This program is free software. It comes without any warranty, to
the extent permitted by applicable law. You can redistribute it
and/or modify it under the terms of the Do What The Fuck You Want
To Public License, Version 2, as published by Sam Hocevar. See
http://sam.zoy.org/wtfpl/COPYING for more details.