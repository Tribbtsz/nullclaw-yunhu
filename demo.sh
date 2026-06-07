#!/bin/bash
# yunhu-channel protocol demo — 验证 JSON-RPC over stdio 协议

set -e
BIN="./yunhu-channel"
TMPDIR="/tmp/yunhu-demo"; mkdir -p "$TMPDIR"; rm -f "$TMPDIR"/*

echo "=== yunhu-channel protocol demo ==="

# 4 条请求一次性 pipe 进去（不含 send，因为需要真实 API）
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"get_manifest","params":{}}' \
  '{"jsonrpc":"2.0","id":2,"method":"start","params":{"runtime":{"name":"yunhu","account_id":"main","state_dir":"/tmp/yunhu-state"},"config":{"token":"demo","webhook_port":18099,"webhook_path":"/webhook/yunhu"}}}' \
  '{"jsonrpc":"2.0","id":9,"method":"stop","params":{}}' \
| "$BIN" > "$TMPDIR/out" 2>"$TMPDIR/log" &
PID=$!

sleep 2
kill $PID 2>/dev/null || true
wait $PID 2>/dev/null || true

echo ""
echo "responses (stdout):"
cat "$TMPDIR/out" | while IFS= read -r line; do
    echo "$line" | python3 -m json.tool 2>/dev/null || echo "$line"
done

echo ""
echo "log (stderr):"
cat "$TMPDIR/log"
