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

Create a file called `.ts3-evict.yaml` inside this folder (e.g. using `nano .ts3-evict.yaml`) and edit it to fit your needs.
See [config](#config) section for examples.

### Config

Example with all options:

```yaml
defaultusername: serveradmin
defaultpassword: your-password-here
defaultqueryport: 10011
defaultports: [9987, 9988]
# timelimit before evicting (in minutes)
timelimit: 5
action: kick
message: Timelimit exceeded.
# delay before doing action (in seconds)
delay: 15
ignoreGroupNames:
  - Server Admin
  - musicbot
servers:
  - IP: ts3.example.com
  - IP: another.ts3.example.com
    ports: [9987]
  - IP: 127.0.0.1
    queryport: 10011
    ports: [9987, 9988, 9989]
    username: serveradmin
    password: your-secret-password
```

Example for a single TS3 server:

```yaml
timelimit: 5 # timelimit before evicting (in minutes)
action: kick
message: Timelimit exceeded.
delay: 15 # delay before doing action (in seconds)
ignoreGroupNames:
  - Server Admin
  - musicbot
servers:
  - IP: 127.0.0.1
    queryport: 10011
    ports: [9987]
    username: serveradmin
    password: your-secret-password
```

## Usage

Run `/opt/ts3-evict/ts3-evict`.

## Build

Run `make`.
