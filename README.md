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

## Development

```bash
make test     # run tests
make lint     # vet
make build    # build binary
make clean    # remove binary
```

## License

MIT
