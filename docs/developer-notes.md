# Developer notes

Here are few notes to myself and possible others to remember what I discovered.

If I mention any code references somewhere it is from decompiled APK of the Andorid version of Nanit mobile client. I used version v2.4.9 for my observations.

## Rest API

Pretty much standard stuff. All can be read from `sources/com/nanit/baby/network/retrofit/RetrofitApi.java`.

There does not seem to be any REST api on local device, only through Nanit servers AFAIK (yet there is still 443 port open which I was not able to access yet).

## Websocket protocol

Websocket seems to be protobuf version 2 (defaults would not work in the way expected in version 3).

Majority of the protocol can be reverse engineered from decompiled `sources/com/nanit/baby/Nanit.java`.

It is possible to connect to the websocket either through Nanit servers or locally. Local websocket runs on port 442 and it is TLS encrypted with self-signed certificate (with wrong CName).

Mobile clients start sending keep-alive packets after 1s and then every 20s. I have not yet experienced connection close with this strategy.

On Nanit servers there are 2 websocket endpoints, 1 for camera (`wss://api.nanit.com/focus/cameras/{camera_uid}/connect`)
and 1 for users (`wss://api.nanit.com/focus/cameras/{camera_uid}/user_connect`). Both seem to be using the same protobuf, but each is accepting different subset of requests.

## Authorization

There seems to be quite mess in request authorization. Probably caused by API being backed by multiple microservices.

The `api.nanit.com` server expects `Authorization: {token}` header.

The `api.nanit.com/focus/*` endpoints expects `Authorization: Bearer {token}`.

The local websocket can be authorized using _User Camera_ token with `Authorization: token {uc_token}`. It can be retrieved from `api.nanit.com/focus/cameras/{camera_uid}/uc_token`. It has longer expiration date so I expect that mobile clients are periodically renewing it so that it is ready for the time they might be offline.

**Warning for local websocket connection:** There seem to be a limit of 2 active connections on the device. The authorization will fail with 403 if you exceed the limit.

## Streaming

Remote streaming is possible through Nanit servers on URL: `rtmps://media-secured.nanit.com/nanit/{baby_uid}.{auth_token}`.

Local streaming seems to be only happening outbound. Meaning you inform cam with the URL (through PUT_STREAMING message) and it starts pushing to that URL a RTMP stream. You can use ie. [nginx-rtmp](https://docs.nginx.com/nginx/admin-guide/dynamic-modules/rtmp/) to accept that stream and restream it however you need (as your own RTMP stream, HLS stream, ...).

## Getting logs

It is possible to retrieve logs from the device using GET_LOGS request (through websocket). They are then sent to the given url using HTTP PUT. The retrieved archive is `tar.gz` (don't let the wrong Content-Type header fool you). After unpacking majority of the interesting stuff is in `journalctl.log`.

In the project there is a HTTP handler prepared for that on `/log` endpoint. It will output the received logs to the `data/logs` folder.

Be aware that getting the logs will take time. In my experience it can even take several minutes for them to arrive. To my understanding, the request might be scheduled for execution given the need to compress the files.

## Events

Motion / sound detection seems to be only distributed through push notifications. I have not intercepted any such message over websocket connection. After further inspection it seems that the Android app is using Intercom push notifications (see `com.nanit.baby.push.fcm.FirebaseMessageHandler`).

https://developers.intercom.com/installing-intercom/docs/android-fcm-push-notifications

I am not sure if it is possible to interface it.

Upon push notification client seems to just fetch the event by ID at `/babies/{baby_uid}/events/{event_uid}`.

Events seems to be listable over `/babies/{baby_uid}/events` but I haven't found this endpoint to be actually used by the mobile app.
