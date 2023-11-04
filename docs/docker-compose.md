# Docker compose quick guide

When running the app as a service, it is useful to use docker-compose for easier configuraiton.

Create `docker-compose.yml` file somewhere on your machine.

Adjust following example (see [.env.sample](../.env.sample) for configuration options).

```yaml
version: '2'
services:
  nanit:
    # Image to pull, adjust the :suffix for your version tag
    image: registry.gitlab.com/adam.stanek/nanit:v0-7
    # Makes the container auto-start whenever you restart your computer
    restart: unless-stopped
    # Expose the RTMP port
    ports:
    - 1935:1935
    # Configuration (see .env.sample file for all the options)
    # Notice: Mind the quotes, whole pairs are quoted instead of just values. If your password contains $ character, replace it with double $$ to avoid interpolation.
    environment:
    - "NANIT_EMAIL=your@email.tld"
    - "NANIT_PASSWORD=XXXXXXXXXXXXX"
    - "NANIT_RTMP_ADDR=xxx.xxx.xxx.xxx:1935"
```

## Control the app container

Run in the same directory as your `docker-compose.yml` file

```bash
# Start the app
docker-compose up -d

# See the logs (Use Ctrl+C to terminate)
docker-compose logs -f nanit

# Stop the app
docker-compose stop

# Upgrade the app (ie. after you have changed the version tag or to pull fresh dev image)
docker-compose pull  # pulls fresh image
docker-compose down  # removes previously created container
docker-compose up -d # creates new container with fresh image (do this after every change in the docker-compose file)

# Uninstall the app
docker-compose down
```
