| :exclamation:  **This software is currently in alpha phase.**   |
|-----------------------------------------------------------------|

<div align="center">
	<a href="https://github.com/pushbits/logo">
		<img height="200px" src="https://raw.githubusercontent.com/pushbits/server/master/.github/logo.png" />
	</a>
</div>

<div align="center">
	<h1>PushBits</h1>
	<h4 align="center">
		Receive your important notifications immediately, over <a href="https://matrix.org/">Matrix</a>.
	</h4>
	<p>PushBits enables you to send push notifications via a simple web API, and delivers them to your users.</p>
</div>

<p align="center">
	<a href="https://github.com/pushbits/server/actions"><img alt="Build status" src="https://img.shields.io/github/workflow/status/pushbits/server/Main"/></a>&nbsp;
	<a href="https://www.pushbits.io/docs/"><img alt="Documentation" src="https://img.shields.io/badge/docs-online-success"/></a>&nbsp;
	<a href="https://www.pushbits.io/api/"><img alt="API Documentation" src="https://img.shields.io/badge/api docs-online-success"/></a>&nbsp;
	<a href="https://matrix.to/#/#pushbits:matrix.org"><img alt="Matrix" src="https://img.shields.io/matrix/pushbits:matrix.org"/></a>&nbsp;
	<!--<a href="https://github.com/pushbits/server/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/release/pushbits/server"/></a>&nbsp;-->
	<a href="https://github.com/pushbits/server/blob/master/LICENSE"><img alt="License" src="https://img.shields.io/github/license/pushbits/server"/></a>
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

This project totally would've used Signal if it would offer a proper API.
Sadly, neither [Signal](https://signal.org/) nor [WhatsApp](https://www.whatsapp.com/) come with an API (at the time of writing) through which PushBits could interact.

In [Telegram](https://telegram.org/) there is an API to run bots, but these are limited in that they cannot create chats by themselves.
If you insist on going with Telegram, have a look at [webhook2telegram](https://github.com/muety/webhook2telegram).

The idea of a federated, synchronized but yet end-to-end encrypted protocol is awesome, but its clients simply aren't really there yet.
Still, if you haven't tried it yet, we'd encourage you to check it out.

## 🤘&nbsp;Features

- [x] Multiple users and multiple channels (applications) per user
- [x] Compatibility with Gotify's API for sending messages
- [x] API and CLI for managing users and applications
- [x] Optional check for weak passwords using [HIBP](https://haveibeenpwned.com/)
- [x] Argon2 as KDF for password storage
- [ ] Two-factor authentication, [issue](https://github.com/pushbits/server/issues/19)
- [ ] Bi-directional key verification, [issue](https://github.com/pushbits/server/issues/20)

## 👮&nbsp;License and Acknowledgments

Please refer to [the LICENSE file](LICENSE) to learn more about the license of this code.
It applies only where not specified differently.

The idea for this software was inspired by [Gotify](https://gotify.net/).

## 💻&nbsp;Development and Contributions

The source code is located on [GitHub](https://github.com/pushbits/server).
You can retrieve it by checking out the repository as follows:
```bash
git clone https://github.com/pushbits/server.git
```

:wrench: **Want to contribute?**
Before moving forward, please refer to [out contribution guidelines](CONTRIBUTING.md).

:mailbox: **Found a security vulnerability?**
Check [this document](SECURITY.md) for information on how you can bring it to our attention.

:star: **Like fancy graphs?** See [our stargazers over time](https://starchart.cc/pushbits/server).
