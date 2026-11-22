#!/usr/bin/env bash
# Linux x64 / macOS arm64 上构建 Go SDK 的测试与全部示例（Windows 用 build.ps1）。
#
#   ./build.sh [-o OUTPUT] [-r RUNTIME_DIR] [-f FULL_PROFILE_RUNTIME_DIR]
#
# OUTPUT 默认 bin/。RUNTIME_DIR 放 libkui_c 与 libkui：默认依次找解包后的 native/lib、对外仓库（git clone kui-go）的
# runtime/<平台>/lib、仓库内开发的 ../c/build（按本版 C++ staging 构建的 sdk/c）。FULL_PROFILE_RUNTIME_DIR 放按目录加载的
# 扩展 / 媒体插件（libkui_ext_* / libkui_media_* / pdfium / FFmpeg / *.filamat）：默认依次找 native/lib、runtime/<平台>/lib、
# 仓库内的 ../cpp/lib。只有 Release 一套：Windows 的 Debug 变体是为了 MSVC Debug CRT。
# 运行测试的 Linux 机器需要显示服务器（没有桌面时用 xvfb-run）。
set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT="$SDK_ROOT/bin"
RUNTIME_DIR=""
FULL_PROFILE_DIR=""
while getopts "o:r:f:" option; do
    case "$option" in
        o) OUTPUT="$OPTARG" ;;
        r) RUNTIME_DIR="$OPTARG" ;;
        f) FULL_PROFILE_DIR="$OPTARG" ;;
        *) echo "usage: $0 [-o OUTPUT] [-r RUNTIME_DIR] [-f FULL_PROFILE_RUNTIME_DIR]" >&2; exit 2 ;;
    esac
done
case "$(uname -s)" in
    Darwin) SUFFIX=dylib; PLATFORM_TAG=macos-arm64 ;;
    *) SUFFIX=so; PLATFORM_TAG=linux-x64 ;;
esac
LIB_C="libkui_c.$SUFFIX"
LIB_CORE="libkui.$SUFFIX"
# 依次找第一个含 $1 的目录（其余参数是候选目录），找到就打印它；都没有返回 1。
first_dir_with() {
    local file="$1" dir
    shift
    for dir in "$@"; do
        if [[ -f "$dir/$file" ]]; then echo "$dir"; return 0; fi
    done
    return 1
}
if [[ -z "$RUNTIME_DIR" ]]; then
    candidates=("$SDK_ROOT/native/lib" "$SDK_ROOT/runtime/$PLATFORM_TAG/lib" "$SDK_ROOT/../c/build")
    RUNTIME_DIR="$(first_dir_with "$LIB_C" "${candidates[@]}")" ||
        { echo "找不到 $LIB_C（可用 -r 指定）；试过：${candidates[*]}" >&2; exit 1; }
fi
if [[ -z "$FULL_PROFILE_DIR" ]]; then
    candidates=("$SDK_ROOT/native/lib" "$SDK_ROOT/runtime/$PLATFORM_TAG/lib" "$SDK_ROOT/../cpp/lib")
    FULL_PROFILE_DIR="$(first_dir_with "libkui_ext_3d.$SUFFIX" "${candidates[@]}")" ||
        { echo "找不到 libkui_ext_3d.$SUFFIX（可用 -f 指定）；试过：${candidates[*]}" >&2; exit 1; }
fi
RUNTIME_DIR="$(cd "$RUNTIME_DIR" && pwd)"
FULL_PROFILE_DIR="$(cd "$FULL_PROFILE_DIR" && pwd)"
for name in "$LIB_C" "$LIB_CORE"; do
    [[ -f "$RUNTIME_DIR/$name" ]] || { echo "缺少运行时：$RUNTIME_DIR/$name" >&2; exit 1; }
done
mkdir -p "$OUTPUT"
OUTPUT="$(cd "$OUTPUT" && pwd)"

export CGO_ENABLED=1
export KUI_C_TEST_RUNTIME="$RUNTIME_DIR/$LIB_C"
cd "$SDK_ROOT"
# canonical 示例数只读包内生成的清单（产品清单 + API 巡览示例），不在脚本里写死。
canonical_count=$(grep -m1 '"count"' examples/canonical/manifest.json | tr -dc '0-9')
[[ -n "$canonical_count" && $canonical_count -gt 0 ]] || { echo "读不到 examples/canonical/manifest.json 的 count" >&2; exit 1; }

go test ./...

quick=0
for dir in examples/*/; do
    name="$(basename "$dir")"
    [[ -f "$dir/main.go" ]] || continue
    go build -trimpath -ldflags "-s -w" -o "$OUTPUT/kui_go_$name" "./examples/$name"
    quick=$((quick + 1))
done
[[ $quick -eq 8 ]] || { echo "Go SDK 应有 8 个专项 examples，实际 $quick" >&2; exit 1; }

sources=$(find examples/canonical/generated -mindepth 2 -maxdepth 2 -name main.go | wc -l | tr -d ' ')
pages=$(find examples/canonical/assets/pages -maxdepth 1 -name '*.html' | wc -l | tr -d ' ')
[[ $sources -eq $canonical_count && $pages -eq $canonical_count ]] || { echo "Go canonical examples 必须是 $canonical_count sources/$canonical_count pages，实际 $sources/$pages" >&2; exit 1; }
canonical_build="$OUTPUT/.canonical-build"
rm -rf "$canonical_build"
mkdir -p "$canonical_build"
go build -trimpath -ldflags "-s -w" -o "$canonical_build/" ./examples/canonical/generated/...
built=0
for exe in "$canonical_build"/*; do
    [[ -f "$exe" && -x "$exe" ]] || continue
    cp -f "$exe" "$OUTPUT/kui_go_canonical_$(basename "$exe")"
    built=$((built + 1))
done
rm -rf "$canonical_build"
[[ $built -eq $canonical_count ]] || { echo "Go canonical examples 应生成 $canonical_count 个程序，实际 $built" >&2; exit 1; }

cp -R examples/canonical/assets "$OUTPUT/"
REPO_ROOT="$(cd "$SDK_ROOT/../.." && pwd)"
if [[ -d "$REPO_ROOT/examples/html_loader/samples" ]]; then
    copy_asset() { [[ -e "$1" ]] || return 0; rm -rf "$2"; mkdir -p "$(dirname "$2")"; cp -R "$1" "$2"; }
    copy_asset "$REPO_ROOT/examples/html_loader/samples" "$OUTPUT/assets/native_samples"
    copy_asset "$REPO_ROOT/tests/script/html" "$OUTPUT/assets/script_html"
    copy_asset "$REPO_ROOT/examples/native/animation/assets/pulse.json" "$OUTPUT/assets/native_animation/pulse.json"
    copy_asset "$REPO_ROOT/resources/audio/sample.mp3" "$OUTPUT/assets/native_media/sample.mp3"
    copy_asset "$REPO_ROOT/resources/video/sample.mp4" "$OUTPUT/assets/native_media/sample.mp4"
    copy_asset "$REPO_ROOT/resources/3d models/stylized_tank.glb" "$OUTPUT/assets/native_3d/stylized_tank.glb"
    copy_asset "$REPO_ROOT/examples/font_awesome_showcase/css" "$OUTPUT/assets/pages/css"
    copy_asset "$REPO_ROOT/examples/font_awesome_showcase/webfonts" "$OUTPUT/assets/pages/webfonts"
fi

# 先铺扩展 / 媒体运行时，再放核心两个库（与 build.ps1 相同顺序）。
for file in "$FULL_PROFILE_DIR"/libkui_ext_*."$SUFFIX" "$FULL_PROFILE_DIR"/libkui_media_*."$SUFFIX" \
            "$FULL_PROFILE_DIR"/libpdfium."$SUFFIX" "$FULL_PROFILE_DIR"/libav* "$FULL_PROFILE_DIR"/libsw* \
            "$FULL_PROFILE_DIR"/LICENSE.FFmpeg*.txt "$FULL_PROFILE_DIR"/*.filamat; do
    [[ -f "$file" ]] && cp -f "$file" "$OUTPUT/"
done
cp -f "$RUNTIME_DIR/$LIB_C" "$RUNTIME_DIR/$LIB_CORE" "$OUTPUT/"
for name in "libkui_ext_3d.$SUFFIX" "libkui_media_audio.$SUFFIX"; do
    [[ -f "$OUTPUT/$name" ]] || { echo "Go bin 运行时门失败：缺少 $name（扩展运行时目录：$FULL_PROFILE_DIR）" >&2; exit 1; }
done
for name in "$LIB_C" "$LIB_CORE"; do
    cmp -s "$RUNTIME_DIR/$name" "$OUTPUT/$name" || { echo "Go bin 核心运行时与 $RUNTIME_DIR 不一致：$name" >&2; exit 1; }
done
echo "Go SDK build PASS: quick=$quick canonical=$built output=$OUTPUT"
