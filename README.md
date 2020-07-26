[![Build status](https://img.shields.io/travis/eikendev/pushbits/master)](https://travis-ci.com/github/eikendev/pushbits/builds/)

## About

PushBits is a relay server for push notifications.
It enables your services to send notifications via a simple web API, and delivers them to you through various messaging services.

The vision is to have compatibility with [Gotify](https://gotify.net/) on the sending side, while on the receiving side established services are used.
This has the advantages that
- plugins written for Gotify and
- clients written for all major platforms can be reused.

For now, only the [Matrix protocol](https://matrix.org/) is supported, but support for different services like [Telegram](https://telegram.org/) could be added in the future.
I am myself experimenting with Matrix currently because I like the idea of a federated, synchronized but still end-to-end encrypted protocol.

The idea for this software and most parts of the initial source are heavily inspired by [Gotify](https://gotify.net/).
Many thanks to [jmattheis](https://jmattheis.de/) for his well-structured code.

## Usage

PushBits is meant to be self-hosted.
You are advised to install PushBits behind a reverse proxy and enable TLS.

At the moment, there is no front-end implemented.
New users and applications need to be created via the API.
Details will be made available once the interface is more stable.

## Development

PushBits is currently in alpha stage.
The API is neither stable, nor is provided functionality guaranteed to work.
Stay tuned! 😉
