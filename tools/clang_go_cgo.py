#!/usr/bin/env python3
"""让 Go cgo 使用 Windows MSVC-target Clang，并保持 KUI 的 MSVC ABI。"""

from __future__ import annotations

import shutil
import subprocess
import sys


def main() -> int:
    clang = shutil.which("clang")
    if not clang:
        print("clang_go_cgo: PATH 中未找到 clang", file=sys.stderr)
        return 127
    # Go 在 Windows 上无条件传入 MinGW 专用 -mthreads；其外部链接器还会
    # 重复传入已由 Go linker 处理的 GNU strip 开关 -s。MSVC-target Clang
    # 不消费这两项，删除它们可避免伪警告，其余参数必须原样保留。
    args = [arg for arg in sys.argv[1:] if arg not in ("-mthreads", "-s")]
    return subprocess.run([clang, *args], check=False).returncode


if __name__ == "__main__":
    raise SystemExit(main())
