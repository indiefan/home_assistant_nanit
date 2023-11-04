# Background

This is a fork of a no-longer-maintained project (https://gitlab.com/adam.stanek/nanit) with added support for Nanit's (now required) 2FA authentication.

# Installation (Docker)

## Pull the Docker Image

While it is possible to build the image locally from the included Dockerfile, it is recommended to install and update by pulling the official image directly from Docker Hub. To pull the image manually without running the container, run:

`docker pull indiefan/nanit`

## Authentication

Because Nanit requires 2FA authentication, before we can start we need to acquire a refresh token for your Nanit account, which can be done by running the included init-nanit.sh CLI tool, which will prompt you for required account information and the 2FA code which will be emailed during the process. The script will save this to a session.json file, where it will be updated automatically going forward. Note that the `/data` volume provided to the script command must be the same used when running the primary container image later.

### Acquire the Refresh Token

Run the bundled init-nanit.sh utility directly via the Docker command line to acquire the token (replace `/path/to/data` with the local path you'd like the container to use for storing session data):

`docker run -it -v /path/to/data:/data --entrypoint=/app/scripts/init-nanit.sh indiefan/nanit`

** Important Note regarding Security**
The refresh token provides complete access to your Nanit account without requiring any additional account information, so be sure to protect your system from access by unauthorized parties, and proceed at your own risk.

## Docker Run

Now that the initial authentication has been done, and the refresh token has been generated, it's time to start the container:

```bash
# Note: use your local IP, reachable from Cam (not 127.0.0.1 nor localhost)

docker run \
  -d \
  --name=nanit \
  --restart unless-stopped \
  -e NANIT_RTMP_ADDR=xxx.xxx.xxx.xxx:1935 \
  -e NANIT_LOG_LEVEL=trace \
  -p 1935:1935 \
  indiefan/nanit:latest
```

If this is your initial run, you may want to omit the `-d` flag so you can observe the output to find your `baby_uid` (which will be needed later if you plan on connecting anything to the feed, like Home Assistant). After getting the baby id (which won't change) you can stop the container and restart it with the `-d` flag.

## Home Assistant

Once the server is running and mirroring the feed, you can then setup an entity in Home Assistant. Open your `configuration.yaml` file and add the following:

```
camera:
- name: Nanit
  platform: ffmpeg
  input: rtmp://xxx.xxx.xxx.xxx:1935/local/[your_baby_uid]
```

Restart Home Assistant and you should now have a camera entity named Nanit for use in dashboards.