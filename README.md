# yunhu-channel

nullclaw external channel plugin for [Yunhu (云湖)](https://www.yhchat.com).

Communicates with nullclaw via JSON-RPC 2.0 over stdin/stdout. Receives Yunhu webhook events and forwards them as inbound messages, and sends outbound messages via Yunhu HTTP API.

## Requirements

- Go 1.21+
- Yunhu Bot Token

## Installation

```bash
git clone https://github.com/yunhu-channel/yunhu-channel.git
cd yunhu-channel
make build
```

Binary will be at `./yunhu-channel`.

## nullclaw Configuration

```json
{
  "channels": {
    "external": [
      {
        "runtime_name": "yunhu",
        "account_id": "main",
        "transport": {
          "command": "/usr/local/bin/yunhu-channel",
          "args": [],
          "timeout_ms": 30000
        },
        "plugin_config_json": "{\"token\":\"YOUR_BOT_TOKEN\",\"webhook_port\":18080,\"webhook_path\":\"/webhook/yunhu\"}"
      }
    ]
  }
}
```

Set the Yunhu Bot webhook URL in the Yunhu console to `https://your-domain.com/webhook/yunhu`.

## Quick Test

**Protocol demo** (no token needed, validates JSON-RPC handshake):
```bash
make build && bash demo.sh
```

**Send a test message** (needs real bot token):
```bash
# 1. Get your Bot Token from https://www.yhchat.com Console
# 2. Get your user/group ID from the Console → Bot → Message log
make build
YUNHU_BOT_TOKEN=your_token bash demo_full.sh your_user_or_group_id
```

**Full end-to-end** (with real phone test):
1. Deploy on a server with a public IP, or use [ngrok](https://ngrok.com) for tunneling
2. Configure nullclaw with the plugin (see above)
3. Set your webhook URL in Yunhu console to `https://your-domain.com/webhook/yunhu`
4. Start nullclaw → plugin starts → webhook server listens → messages flow

## Development

```bash
make test     # run tests
make lint     # vet
make build    # build binary
make clean    # remove binary
```

## License

MIT
