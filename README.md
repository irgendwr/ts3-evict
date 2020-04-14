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
curl -O -L https://github.com/irgendwr/ts3-evict/releases/latest/download/ts3-evict_Linux_x86_64.tar.gz
tar -xvzf ts3-evict_Linux_x86_64.tar.gz
```

Create a file called `.ts3-evict.yaml` inside this folder (e.g. using `nano .ts3-evict.yaml`) and edit it to fit your needs.
See [config](#config) section for examples.

### Config

**Note: Do not use Tabs! Indent config with spaces instead.**

Example with all options:

```yaml
DefaultUsername: serveradmin
DefaultPassword: your-password-here
DefaultQueryPort: 10011
DefaultPorts: [9987, 9988]
Violators: violators.csv
# Timelimit (in minutes) before eviction 
Timelimit: 5
Kicklimit: 3
Action: kick
Message: Timelimit exceeded.
KickMessage: Timelimit exceeded.
BanMessage: Timelimit exceeded.
# Delay (in seconds) before doing action
Delay: 5
IgnoreGroupNames:
  - Server Admin
  - Server Query Admin
  - musicbot
servers:
  - IP: ts3.example.com
  - IP: another.ts3.example.com
    Ports: [9987]
  - IP: 127.0.0.1
    QueryPort: 10011
    Ports: [9987, 9988, 9989]
    Username: serveradmin
    Password: your-secret-password
```

Example for a single TS3 server:

```yaml
Timelimit: 5 # Timelimit (in minutes) before eviction 
Action: kick
Message: Timelimit exceeded.
Delay: 5 # Delay (in seconds) before doing action
IgnoreGroupNames:
  - Server Admin
  - Server Query Admin
  - musicbot
Servers:
  - IP: 127.0.0.1
    QueryPort: 10011
    Ports: [9987]
    Username: serveradmin
    Password: your-secret-password
```

## Usage

Run `/opt/ts3-evict/ts3-evict`.

## Build

Run `make`.
