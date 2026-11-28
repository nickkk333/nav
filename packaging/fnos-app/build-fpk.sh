#!/usr/bin/env bash
# ============================================================
# 打包飞牛OS（fnOS）.fpk 安装包（原生模式，不使用 Docker）
#
# 用法：
#   bash packaging/fnos-app/build-fpk.sh [version] [platform] [output]
#     version  应用版本，写进 manifest 与二进制版本号（默认 dev）
#     platform 目标平台，x86 或 arm（默认 x86）
#     output   产物路径，相对路径基于仓库根目录（默认 bin/nav-fnos-amd64.fpk）
#
# 依赖：bash / git / tar / sed / md5sum / go（Linux、macOS、WSL、Git Bash 均可）
# 前置：前端产物 public/ 已构建（make ui-build）；缺失时会直接报错退出。
# 说明：生命周期脚本由第三方通用打包框架 conversun/fnos-apps 提供，这里只做
#       应用源组装 + app.tgz + 调用其 build-fpk.sh。
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APP_SRC="$SCRIPT_DIR"

VERSION="${1:-dev}"
PLATFORM="${2:-x86}"

case "$PLATFORM" in
  x86) GOARCH="amd64"; DEFAULT_OUTPUT="bin/nav-fnos-amd64.fpk" ;;
  arm) GOARCH="arm64"; DEFAULT_OUTPUT="bin/nav-fnos-arm64.fpk" ;;
  *) echo "不支持的平台：${PLATFORM}（可选 x86 / arm）" >&2; exit 1 ;;
esac

OUTPUT="${3:-$DEFAULT_OUTPUT}"
case "$OUTPUT" in
  /*|[A-Za-z]:[\\/]*) ;;
  *) OUTPUT="$REPO_ROOT/$OUTPUT" ;;
esac

# go:embed 需要前端产物；这里不替用户跑 npm，避免打包脚本偷偷安装依赖
if [ ! -f "$REPO_ROOT/public/index.html" ]; then
  echo "缺少前端产物 public/index.html，请先执行：make ui-build" >&2
  exit 1
fi

# 编译目标架构的静态二进制；总是重新编译，避免复用到旧前端产物
# （需要复用现成二进制时用 NAV_BINARY=/path/to/nav 指定）
if [ -n "${NAV_BINARY:-}" ]; then
  BINARY="$NAV_BINARY"
  if [ ! -f "$BINARY" ]; then
    echo "NAV_BINARY 指定的文件不存在：$BINARY" >&2
    exit 1
  fi
else
  BINARY="$REPO_ROOT/bin/nav-linux-${GOARCH}"
  echo ">>> 编译 linux/${GOARCH} 二进制"
  mkdir -p "$(dirname "$BINARY")"
  COMMIT="$(git -C "$REPO_ROOT" rev-parse --short HEAD 2>/dev/null || echo none)"
  ( cd "$REPO_ROOT" && CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" \
      go build -trimpath -tags timetzdata \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o "$BINARY" . )
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT
WORK_DIR="$TMP_ROOT/work"
mkdir -p "$WORK_DIR"

# app.tgz 会解压到应用的 target 目录（TRIM_APPDEST）：
#   nav            主程序（内嵌前端）
#   bin/nav-server 启动包装（校验数据目录可写后再启动）
#   ui/            桌面图标与应用入口配置
echo ">>> 组装 app.tgz"
APP_ROOT="$WORK_DIR/app"
mkdir -p "$APP_ROOT/bin"
cp "$BINARY" "$APP_ROOT/nav"
chmod +x "$APP_ROOT/nav"
# 开源许可与来源说明（MIT 要求随分发保留版权与许可声明）
cp "$REPO_ROOT/LICENSE" "$APP_ROOT/LICENSE"
cp "$REPO_ROOT/NOTICE" "$APP_ROOT/NOTICE"
cp "$APP_SRC/bin/nav-server" "$APP_ROOT/bin/nav-server"
chmod +x "$APP_ROOT/bin/nav-server"
cp -a "$APP_SRC/ui" "$APP_ROOT/ui"
( cd "$APP_ROOT" && tar -czf "$WORK_DIR/app.tgz" . )

TOOLKIT_REPO="https://github.com/conversun/fnos-apps"
TOOLKIT_DIR="$TMP_ROOT/fnos-toolkit"
APP_DIR="$TOOLKIT_DIR/apps/nav"

echo ">>> 获取 fnOS 打包框架"
if [ -n "${FNOS_TOOLKIT_DIR:-}" ] && [ -d "${FNOS_TOOLKIT_DIR}/scripts" ]; then
  # 本地已有 checkout 时直接复用（离线构建 / 调试打包脚本）
  echo "    使用本地框架：${FNOS_TOOLKIT_DIR}"
  cp -a "$FNOS_TOOLKIT_DIR" "$TOOLKIT_DIR"
else
  git clone --depth 1 "$TOOLKIT_REPO" "$TOOLKIT_DIR"
fi

echo ">>> 组装应用目录"
mkdir -p "$APP_DIR/fnos"
cp -a "$APP_SRC"/. "$APP_DIR/fnos/"
rm -f "$APP_DIR/fnos/build-fpk.sh"
# ICON.PNG(128) / ICON_256.PNG(256) 已经在 packaging/fnos-app 下，随上面的 cp -a 一起带过去

echo ">>> 打包 fpk"
mkdir -p "$(dirname "$OUTPUT")"
( cd "$TOOLKIT_DIR" \
  && NAME="$(scripts/build-fpk.sh apps/nav "$WORK_DIR/app.tgz" "$VERSION" "$PLATFORM" | tail -n 1)" \
  && cp "$NAME" "$OUTPUT" )

echo ">>> 完成：$OUTPUT"
ls -lh "$OUTPUT"
