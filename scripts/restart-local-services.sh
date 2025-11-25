#!/bin/bash

# Zervigo 本地服务重启脚本

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

STOP_SCRIPT="$PROJECT_ROOT/scripts/stop-local-services.sh"
START_SCRIPT="$PROJECT_ROOT/scripts/start-local-services.sh"

if [ ! -x "$STOP_SCRIPT" ] || [ ! -x "$START_SCRIPT" ]; then
    echo "❌ 未找到启动/停止脚本，请确认项目结构是否正确"
    exit 1
fi

echo "🔄 正在重启 Zervigo 本地开发环境..."

"$STOP_SCRIPT"
"$START_SCRIPT"

echo "✅ 重启完成"

