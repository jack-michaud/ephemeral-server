# Development

## Building in Docker Image

```sh
# A build arg is required. 
# linux/amd64, linux/arm64 will cover most development options
# This will install ansible, terraform, and other dependencies.
# It will also build the go project. The run command below will start the bot.
docker build --build-arg="TARGETPLATFORM=linux/amd64" -t ephemeralbot .
```

```
# -e .*  Puts the environment variables into the docker container. 
# -e SECRET_KEY is an AES-256 key used to encrypt configs.
# -e DISCORD_CLIENT_SECRET is a Discord *bot* access key (not client_id/client_secret)
# --net=host will give the bot access to the host's network interface. Useful because I run consul
#            on the host currently.
docker run -e DISCORD_CLIENT_SECRET -e SECRET_KEY --net=host -it ephemeralbot
```

## Consul

Consul is required. The bot looks for consul on 127.0.0.1:8500.

