# ts3-evict

![Build status](https://github.com/irgendwr/ts3-evict/workflows/build/badge.svg)
![Release status](https://github.com/irgendwr/ts3-evict/workflows/release/badge.svg)
[![GitHub Release](https://img.shields.io/github/release/irgendwr/ts3-evict.svg)](https://github.com/irgendwr/ts3-evict/releases)

Evict users/clients from a TeamSpeak 3 server after a given time; useful for demo servers.

## Installation

### Linux

Download the latest release into `/opt/ts3-evict/` (or any other folder):

```bash
mkdir -p /opt/ts3-evict/
cd /opt/ts3-evict/
curl -O -L https://github.com/irgendwr/sinusbot-spotify/releases/latest/download/ts3-evict_Linux_x86_64.tar.gz
tar -xvzf ts3-evict_Linux_x86_64.tar.gz
```

Create a file called `.ts3-evict.yaml` inside this folder (e.g. using `nano .ts3-evict.yaml`) and edit it with your spotify credentials:

```yaml
defaultusername: serveradmin
defaultpassword: your-password-here
defaultqueryport: 10011
defaultport: 9987
# timelimit in minutes
timelimit: 5
action: kick
message: Timelimit exceeded.
# delay in seconds
delay: 15
ignoreGroupNames:
  - Server Admin
  - musicbot
servers:
  - IP: boegle.me
```

## Usage

Run `/opt/ts3-evict/ts3-evict`.

## Build

Run `make`.
