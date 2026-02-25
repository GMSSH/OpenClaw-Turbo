#!/bin/bash

# ====================================================
# 脚本名称: install_node.sh
# 功能: 智能识别系统环境，自动匹配最高可用 Node 版本
# 适配: CentOS (7/Stream), Ubuntu (18+), Debian (9+)
# ====================================================

set -e

# 1. 基础系统信息检测
OS_TYPE=$(lsb_release -is 2>/dev/null || cat /etc/os-release | grep -w "ID" | cut -d= -f2 | tr -d '"')
OS_VER=$(lsb_release -rs 2>/dev/null || cat /etc/os-release | grep -w "VERSION_ID" | cut -d= -f2 | tr -d '"')
GLIBC_VER=$(ldd --version | head -n1 | grep -oE '[0-9]+\.[0-9]+' | tail -n1)

echo "🔍 检测到系统: $OS_TYPE $OS_VER (glibc: $GLIBC_VER)"

# 2. 版本决策逻辑
if [ "$(echo "$GLIBC_VER < 2.25" | bc -l)" -eq 1 ]; then
    TARGET_NODE="16"
    echo "⚠️ 系统 glibc 版本过低，为保证稳定性，将安装 Node $TARGET_NODE"
elif [ "$(echo "$GLIBC_VER < 2.28" | bc -l)" -eq 1 ]; then
    TARGET_NODE="18"
    echo "ℹ️ 系统环境较老，将安装 Node $TARGET_NODE"
else
    TARGET_NODE="24"
    echo "🚀 系统环境良好，将安装最新的 Node $TARGET_NODE"
fi

# 3. 安装基础依赖
echo "[1/6] 正在安装基础依赖..."
if command -v apt-get >/dev/null 2>&1; then
    apt-get update && apt-get install -y git curl sudo tar bc
elif command -v dnf >/dev/null 2>&1; then
    dnf install -y git curl sudo tar bc
elif command -v yum >/dev/null 2>&1; then
    yum install -y git curl sudo tar bc
fi

# 4. 全局 NVM 安装
echo "[2/6] 正在配置系统级 NVM (Gitee 镜像)..."
export NVM_DIR="/usr/local/nvm"
mkdir -p $NVM_DIR
if [ ! -d "$NVM_DIR/.git" ]; then
    git clone https://gitee.com/mirrors/nvm.git "$NVM_DIR"
    cd "$NVM_DIR" && git checkout v0.40.1
fi

# 5. 写入全局环境变量
echo "[3/6] 写入全局 profile 配置..."
cat << EOF > /etc/profile.d/gmssh_node.sh
export NVM_DIR="/usr/local/nvm"
[ -s "\$NVM_DIR/nvm.sh" ] && \. "\$NVM_DIR/nvm.sh"
[ -s "\$NVM_DIR/bash_completion" ] && \. "\$NVM_DIR/bash_completion"
export PNPM_HOME="/usr/local/share/pnpm"
export PATH="\$PNPM_HOME:\$PATH"
EOF
source /etc/profile.d/gmssh_node.sh

# 6. 安装目标版本 Node
echo "[4/6] 正在通过镜像安装 Node $TARGET_NODE..."
export NVM_NODEJS_ORG_MIRROR=https://npmmirror.com/mirrors/node/
nvm install $TARGET_NODE
nvm alias default $TARGET_NODE
nvm use default

# 7. pnpm 加速配置
echo "[5/6] 正在配置 pnpm 加速..."
yes | npm config set registry https://registry.npmmirror.com
corepack enable || npm install -g pnpm@latest
yes | pnpm config set registry https://registry.npmmirror.com

# 8. 建立软链接
echo "[6/6] 建立二进制软链接..."
ln -sf $(which node) /usr/bin/node
ln -sf $(which pnpm) /usr/bin/pnpm

echo "------------------------------------------------"
echo "✅ 安装完成！当前环境: Node $(node -v)"
if [ "$TARGET_NODE" -lt "24" ]; then
    echo "❌ 注意: 当前系统由于内核组件限制，无法原生运行 OpenClaw (要求 Node 24+)。"
    echo "建议: 升级系统至 Ubuntu 20.04+ 或使用 Docker 部署。"
fi
echo "------------------------------------------------"
