---
name: virtualtext
description: Use the VirtualText vt CLI to monitor response time, observations, and sentiment.
---

# VirtualText CLI

If `vt` is missing:

```bash
curl -fsSL https://raw.githubusercontent.com/VirtualPBX/virtualtext-cli/main/install.sh | bash
```

If `vt auth status` shows no token, ask the human to run:

```bash
vt auth login --host https://THEIR_INSTANCE.virtualtext.app
```

Then:

```bash
vt metrics response --since 24h
vt observations list --status pending
vt conversations list --waiting
```

Do not clone the Rails app. The CLI talks to `/api/ops` with an ops-only API key.
