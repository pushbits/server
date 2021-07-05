| :exclamation:  **This software is currently in alpha phase.**   |
|-----------------------------------------------------------------|

<div align="center">
	<h1>PushBits</h1>
	<h4 align="center">
		Receive your important notifications immediately, over <a href="https://matrix.org/">Matrix</a>.
	</h4>
	<p>PushBits enables you to send push notifications via a simple web API, and delivers them to your users.</p>
</div>

<p align="center">
	<a href="https://github.com/pushbits/server/actions"><img alt="Build status" src="https://img.shields.io/github/workflow/status/pushbits/server/Main"/></a>&nbsp;
	<a href="https://hub.docker.com/r/eikendev/pushbits"><img alt="Downloads" src="https://img.shields.io/docker/pulls/eikendev/pushbits"/></a>&nbsp;
	<a href="https://hub.docker.com/r/eikendev/pushbits"><img alt="Image size" src="https://img.shields.io/docker/image-size/eikendev/pushbits"/></a>&nbsp;
	<a href="https://github.com/pushbits/server/blob/master/LICENSE"><img alt="License" src="https://img.shields.io/github/license/pushbits/server"/></a>&nbsp;
</p>

## 💡&nbsp;About

PushBits is a relay server for push notifications.
It enables you to send notifications via a simple web API, and delivers them to you through [Matrix](https://matrix.org/).
This is similar to what [Pushover](https://pushover.net/) and [Gotify](https://gotify.net/) offer, but it does not require an additional app.

The vision is to have compatibility with Gotify on the sending side, while on the receiving side an established service is used.
This has the advantages that
- sending plugins written for Gotify (like those for [Watchtower](https://containrrr.dev/watchtower/) and [Jellyfin](https://jellyfin.org/)) as well as
- receiving clients written for Matrix
can be reused.

### Why Matrix instead of X?

I would totally do this with Signal if there was a proper API.
Unfortunately, neither [Signal](https://signal.org/) nor [WhatsApp](https://www.whatsapp.com/) come with an API through which PushBits could interact.

In [Telegram](https://telegram.org/) there is an API to run bots, but these are limited in that they cannot create chats by themselves.
If you insist on going with Telegram, have a look at [webhook2telegram](https://github.com/muety/webhook2telegram).

I myself started using Matrix only for this project.
The idea of a federated, synchronized but yet end-to-end encrypted protocol is awesome, but its clients simply aren't really there yet.
Still, if you haven't tried it yet, I suggest you to check it out.

## 🤘&nbsp;Features

- [x] Multiple users and multiple channels (applications) per user
- [x] Compatibility with Gotify's API for sending messages
- [x] API and CLI for managing users and applications
- [x] Optional check for weak passwords using [HIBP](https://haveibeenpwned.com/)
- [x] Argon2 as KDF for password storage
- [ ] Two-factor authentication, [issue](https://github.com/pushbits/server/issues/19)
- [ ] Bi-directional key verification, [issue](https://github.com/pushbits/server/issues/20)

## 🚀&nbsp;Installation

PushBits is meant to be self-hosted.
That means you have to install it on your own server.

Currently, the only supported way of installing PushBits is via [Docker](https://www.docker.com/) or [Podman](https://podman.io/).
The image is hosted [here on Docker Hub](https://hub.docker.com/r/eikendev/pushbits).

| :warning:  **You are advised to install PushBits behind a reverse proxy and enable TLS.** Otherwise, your credentials will be transmitted unencrypted.   |
|----------------------------------------------------------------------------------------------------------------------------------------------------------|

## ⚙&nbsp;Configuration

To see what can be configured, have a look at the `config.sample.yml` file inside the root of the repository.

Settings can optionally be provided via environment variables.
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

## 📄&nbsp;Usage

Now, how can you interact with the server?
I wrote [a little CLI tool called pbcli](https://github.com/PushBits/cli) to make basic API requests to the server.
It helps you to create new users and applications.
You will find further instructions in the linked repository.

At the time of writing, there is no fancy GUI built-in, and I'm not sure if this is necessary at all.
I don't do much front end development myself, so if you want to contribute in this regard I'm happy if you reach out!

After you have created a user and an application, you can use the API to send a push notification to your Matrix account.

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	--data '{"message":"my message","title":"my title"}' \
	"https://pushbits.example.com/message?token=$PB_TOKEN"
```

Note that the token is associated with your application and has to be kept secret.
You can retrieve the token using [pbcli](https://github.com/PushBits/cli) by running following command.

```bash
pbcli application show $PB_APPLICATION --url https://pushbits.example.com --username $PB_USERNAME
```

### Message options

Messages are supporting three different syntaxes:

* text/plain
* text/html
* text/markdown

To set a specific syntax you need to set the `extras` ([inspired by Gotifys message extras](https://gotify.net/docs/msgextras#clientdisplay)):

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	--data '{"message":"my message with\n\n**Markdown** _support_.","title":"my title","extras":{"client::display":{"contentType": "text/markdown"}}}' \
	"https://pushbits.example.com/message?token=$PB_TOKEN"
```

HTML-Content might not be fully rendered in your Matrix-Client - see the corresponding [Matrix specs](https://spec.matrix.org/unstable/client-server-api/#mroommessage-msgtypes). This also holds for Markdown, as it is transfered to the corresponding HTML-syntax.

### Deleting a Message

You can delete a message, this will send a notification in response to the original message informing you that the message is "deleted". 

You need the message ID for deleting a message. As it might contain characters not valid in uris we provide an additional `id_url_encoded` field for messages, use that value for deleting a message.

```bash
curl \
	--request DELETE \
	"https://pushbits.example.com/message/${MESSAGE_ID}?token=$PB_TOKEN"
```

## 👮&nbsp;Acknowledgments

The idea for this software and most parts of the initial source are heavily inspired by [Gotify](https://gotify.net/).
Many thanks to [jmattheis](https://jmattheis.de/) for his well-structured code.

## 💻&nbsp;Development

The source code is located on [GitHub](https://github.com/pushbits/server).
You can retrieve it by checking out the repository as follows.

```bash
git clone https://github.com/pushbits/server.git
```

[![Stargazers over time](https://starchart.cc/pushbits/server.svg)](https://starchart.cc/pushbits/server)

### Testing

Testing is essential for deliviering a good and reliable software. PushBits uses golangs integrated test features. Unfortunately writing tests is quite time consuming and therefore not every feature and every line of code is automatically tested. Feel free to help us improve our tests.

To run tests for a single (sub)module you can simply execute the following command in the modules folder.

```bash
go test
```

To get the testing coverage for a module use the `-cover` flag. 

```bash
go test -cover
```

To execute a single test use the `-run` flag.

```bash
go test -run "TestApi_getUser"
```

Running tests for all PushBits module is done like this:

```bash
make test
```