# VirtualText CLI (`vt`)

Go CLI for VirtualText ops: response time, observations, sentiment, conversations.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/VirtualPBX/virtualtext-cli/main/install.sh | bash
```

## Login (creates an ops-only API key)

```bash
vt auth login --host https://use.virtualtext.app
```

A browser window asks an admin to Allow. The token is stored in `~/.config/virtualtext/config.yaml` (mode 0600).

Already have a key?

```bash
vt auth login --host https://use.virtualtext.app --token YOUR_KEY
```

## Use

```bash
vt metrics response --since 24h --kind human
vt observations list --status pending
vt conversations list --waiting
vt skill install
```

Bots: install the skill, then call `vt`. Do not clone the Rails app.
