#!/bin/bash
# yunhu-channel 完整 demo — 需要真实 YUNHU_BOT_TOKEN
# 用法: YUNHU_BOT_TOKEN=your_token bash demo_full.sh

set -e
BIN="./yunhu-channel"
TMPDIR="/tmp/yunhu-demo"; mkdir -p "$TMPDIR"; rm -f "$TMPDIR"/*

if [ -z "$YUNHU_BOT_TOKEN" ]; then
    echo "请设置 YUNHU_BOT_TOKEN 环境变量"
    echo "用法: YUNHU_BOT_TOKEN=your_token bash demo_full.sh"
    exit 1
fi

TARGET="${1:-user_or_group_id}"
if [ "$TARGET" = "user_or_group_id" ]; then
    echo "请提供目标 ID"
    echo "用法: YUNHU_BOT_TOKEN=xxx bash demo_full.sh <目标ID>"
    exit 1
fi

echo "=== full demo: send message via yunhu API ==="
echo "target: $TARGET"

printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"get_manifest","params":{}}' \
  "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"start\",\"params\":{\"runtime\":{\"name\":\"yunhu\",\"account_id\":\"main\",\"state_dir\":\"$TMPDIR/state\"},\"config\":{\"token\":\"$YUNHU_BOT_TOKEN\",\"webhook_port\":18099,\"webhook_path\":\"/webhook/yunhu\"}}}" \
  "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"send\",\"params\":{\"runtime\":{\"name\":\"yunhu\",\"account_id\":\"main\"},\"message\":{\"target\":\"$TARGET\",\"text\":\"Hello from yunhu-channel demo 🎉\",\"stage\":\"final\",\"media\":[]}}}" \
  "{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"send_rich\",\"params\":{\"runtime\":{\"name\":\"yunhu\",\"account_id\":\"main\"},\"message\":{\"target\":\"$TARGET\",\"text\":\"请选择一项\",\"attachments\":[],\"choices\":[{\"id\":\"a\",\"label\":\"选项 A\",\"submit_text\":\"A\"},{\"id\":\"b\",\"label\":\"选项 B\",\"submit_text\":\"B\"}]}}}" \
  '{"jsonrpc":"2.0","id":9,"method":"stop","params":{}}' \
| "$BIN" > "$TMPDIR/out" 2>"$TMPDIR/log" &
PID=$!

sleep 8
kill $PID 2>/dev/null || true
wait $PID 2>/dev/null || true

echo ""
echo "stdout:"
cat "$TMPDIR/out" | while IFS= read -r line; do
    echo "$line" | python3 -m json.tool
done

echo ""
echo "log:"
cat "$TMPDIR/log"
