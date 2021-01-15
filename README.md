[![Build status](https://img.shields.io/github/workflow/status/eikendev/pushbits/Main)](https://github.com/eikendev/pushbits/actions)
[![Docker Pulls](https://img.shields.io/docker/pulls/eikendev/pushbits)](https://hub.docker.com/r/eikendev/pushbits)
![License](https://img.shields.io/github/license/eikendev/pushbits)

# PushBits

| :exclamation:  This software is currently in alpha phase.   |
|-------------------------------------------------------------|

## About

PushBits is a relay server for push notifications.
It enables your services to send notifications via a simple web API, and delivers them to you through [Matrix](https://matrix.org/).
This is similar to what [PushBullet](https://www.pushbullet.com/), [Pushover](https://pushover.net/), and [Gotify](https://gotify.net/) offer, but a lot less complex.

The vision is to have compatibility with Gotify on the sending side, while on the receiving side an established service is used.
This has the advantages that
- sending plugins written for Gotify (like those for [Watchtower](https://containrrr.dev/watchtower/) and [Jellyfin](https://jellyfin.org/)) as well as
- receiving clients written for the messaging service
can be reused.

### Why Matrix instead of X?

For now, only Matrix is supported, but support for different services like [Telegram](https://telegram.org/) could be added in the future.
[WhatsApp](https://www.whatsapp.com/) and [Signal](https://signal.org/) unfortunately do not have an API through which PushBits can interact.

I am myself experimenting with Matrix currently because I like the idea of a federated, synchronized but still end-to-end encrypted protocol.
If you haven't tried it yet, I suggest you to check it out.

## Configuration

PushBits is meant to be self-hosted.
You are advised to install PushBits behind a reverse proxy and enable TLS.

To see what can be configured, have a look at the `config.sample.yml` file inside the root of the repository.

Settings can optionally be provided via the environment.
The name of the environment variable is composed of a starting `PUSHBITS_`, followed by the keys of the setting, all
joined with `_`.
As an example, the HTTP port can be provided as an environment variable called `PUSHBITS_HTTP_PORT`.

To get started, here is a Docker Compose file you can use.
```yaml
version: '2'

services:
    server:
        image: eikendev/pushbits:latest
        ports:
            - 8080:8080
        environment:
            PUSHBITS_DATABASE_DIALECT: 'sqlite3'
            PUSHBITS_ADMIN_MATRIXID: '@your/matrix/username:matrix.org' # The Matrix account on which the admin will receive their notifications.
            PUSHBITS_ADMIN_PASSWORD: 'your/pushbits/password' # The login password of the admin account. Default username is 'admin'.
            PUSHBITS_MATRIX_USERNAME: 'your/matrix/username' # The Matrix account from which notifications are sent to all users.
            PUSHBITS_MATRIX_PASSWORD: 'your/matrix/password' # The password of the above account.
        volumes:
            - /etc/localtime:/etc/localtime:ro
            - /etc/timezone:/etc/timezone:ro
            - ./data:/data
```

In this example, the configuration file would be located at `./data/config.yml` on the host.
The SQLite database would be written to `./data/pushbits.db`.
**Don't forget to adjust the permissions** of the `./data` directory, otherwise PushBits will fail to operate.

## Usage

Now, how can you interact with the server?
At the time of writing, there is no fancy GUI built-in.
I don't do much front end development myself, so if you want to contribute in this regard I'm happy if you reach out!

Anyway, I wrote [a little CLI tool](https://github.com/PushBits/cli) to make basic API requests to the server.
It helps you to create new users and applications.
You will find further instructions in the linked repository.

After you have setup a user and an application, you can use the API to send a push notification to your Matrix account.

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	--data '{"message":"my message","title":"my title"}' \
	"https://pushbits.example.com/message?token=$TOKEN"
```

## Acknowledgments

The idea for this software and most parts of the initial source are heavily inspired by [Gotify](https://gotify.net/).
Many thanks to [jmattheis](https://jmattheis.de/) for his well-structured code.

## Development

The source code is located on [GitHub](https://github.com/eikendev/pushbits).
You can retrieve it by checking out the repository as follows.

```bash
git clone https://github.com/eikendev/pushbits.git
```
