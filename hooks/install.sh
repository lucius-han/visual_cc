#!/bin/bash
# visual_cc hook install script
# Prints Claude Code settings.json hook configuration for manual merging

HOOK_BIN="$(cd "$(dirname "$0")/.." && pwd)/visual_cc-hook"

if [ ! -f "$HOOK_BIN" ]; then
    echo "Error: visual_cc-hook binary not found at $HOOK_BIN"
    echo "Build it first:"
    echo "  go build -o visual_cc-hook ./cmd/visual_cc-hook/"
    exit 1
fi

# S12: warn if the path contains spaces (would produce invalid JSON)
if [[ "$HOOK_BIN" == *" "* ]]; then
    echo "Warning: the binary path contains spaces: $HOOK_BIN"
    echo "Move the binary to a path without spaces before registering the hook."
    echo "Example: sudo cp visual_cc-hook /usr/local/bin/visual_cc-hook"
    echo "Then use '/usr/local/bin/visual_cc-hook' as the command below."
    exit 1
fi

SETTINGS="$HOME/.claude/settings.json"

echo "Add the following to the 'hooks' section of $SETTINGS"
echo "(merge manually if hooks already exist):"
echo ""
cat <<EOF
{
  "hooks": {
    "PreToolUse": [
      { "hooks": [{ "type": "command", "command": "$HOOK_BIN" }] }
    ],
    "PostToolUse": [
      { "hooks": [{ "type": "command", "command": "$HOOK_BIN" }] }
    ],
    "Stop": [
      { "hooks": [{ "type": "command", "command": "$HOOK_BIN" }] }
    ],
    "Notification": [
      { "hooks": [{ "type": "command", "command": "$HOOK_BIN" }] }
    ]
  }
}
EOF
