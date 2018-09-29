# Steam Gameserver REST API

A REST API for pulling Steam Gameserver Tokens through Steamworks Web API.

Its wraps the IGameServersService Interface, and the code has been built on knowledge from two sources.  
A [community made API reference](http://steamwebapi.azurewebsites.net/).  
And the [Steamworks Documentation Website](https://partner.steamgames.com/doc/webapi/IGameServersService).

## Getting started

The application need a `STEAM_WEB_API_KEY` environment variable, which can be generated / found [here](https://steamcommunity.com/dev/apikey).

It will listen on `0.0.0.0:8000`, unless you override with the `STEAM_WEB_API_BIND_ADDRESS` environment variable.

It returns tokens as text/plain on the following URL:

> [GET] /token/{appID}/{memo}

* **appID** is the Steam Application ID (e.g. 740 for CSGO dedicated server)
* **memo** is a note that uniquely identifies a gameserver

The library it uses to communicate with Steamworks Web API is [nested in this project](steam/README.md).

## Errors

Errors from the Steamworks Web API will be forwarded as JSON objects.

> { "error": "some error happened" }

## Build

```sh
# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o steam-api.exe main.go app.go

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o steam-api main.go app.go

# OSX
GOOS=darwin go build -ldflags="-s -w" -o steam-api main.go app.go
```

Optionally, you can cut down binary size with `upx --brute`.
