#!/usr/bin/env python3
"""将 CodeBuddy account/auth JSON 转换为指定的 token JSON（仅使用标准库）。

用法：
    python convert_codebuddy.py input.json output.json
    python convert_codebuddy.py input.json
    python convert_codebuddy.py input.json output.json --expires-in 2419200
    python convert_codebuddy.py --dir ./input ./output

目录模式只处理输入目录下的 JSON 文件，不递归子目录；输出保留原文件名。
"""

import argparse
import json
import sys
import time
from datetime import datetime, timezone
from pathlib import Path


def convert(data, expires_in=2419200):
    if not isinstance(data, dict) or not isinstance(data.get("auth"), dict):
        raise ValueError("输入 JSON 必须包含 auth 对象")
    auth = data["auth"]
    for key in ("accessToken", "refreshToken"):
        if not isinstance(auth.get(key), str) or not auth[key].strip():
            raise ValueError(f"auth.{key} 必须是非空字符串")

    # 源 expiresAt 是 Unix 秒时间戳；目标 expired 使用 UTC 的 ISO 8601 格式。
    raw_expiry = auth.get("expiresAt")
    if isinstance(raw_expiry, bool) or not isinstance(raw_expiry, (int, str)):
        raise ValueError("auth.expiresAt 必须是 Unix 秒时间戳（整数）")
    try:
        expiry = int(raw_expiry)
        expired = datetime.fromtimestamp(expiry, timezone.utc).strftime(
            "%Y-%m-%dT%H:%M:%SZ"
        )
    except (ValueError, OverflowError, OSError):
        raise ValueError("auth.expiresAt 不是有效的 Unix 秒时间戳") from None

    return {
        "access_token": auth["accessToken"],
        "base_url": "https://copilot.tencent.com/v2",
        "disabled": False,
        "expired": expired,
        # 按示例默认填写 28 天，不代表从转换时起的剩余有效期。
        "expires_in": expires_in,
        "note": "",
        "prefix": "",
        "refresh_token": auth["refreshToken"],
        "timestamp": time.time_ns() // 1_000_000,
        "token_type": "Bearer",
        "type": "codebuddy-cn",
    }


def convert_file(source, output, expires_in):
    if source.resolve() == output.resolve():
        raise ValueError("输入和输出不能是同一个文件")
    # 兼容 Windows 编辑器生成的 UTF-8 BOM。
    data = json.loads(source.read_text(encoding="utf-8-sig"))
    result = convert(data, expires_in=expires_in)
    output.write_text(
        json.dumps(result, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


def convert_directory(source, output, expires_in):
    if not source.is_dir():
        raise ValueError("输入路径不是有效目录")
    if source.resolve() == output.resolve():
        raise ValueError("输入目录和输出目录不能相同")
    files = sorted(
        path for path in source.iterdir()
        if path.is_file() and path.suffix.lower() == ".json"
    )
    if not files:
        raise ValueError("输入目录中没有 JSON 文件")
    output.mkdir(parents=True, exist_ok=True)

    failed = 0
    for path in files:
        destination = output / path.name
        try:
            convert_file(path, destination, expires_in)
        except (OSError, UnicodeError, ValueError) as exc:
            failed += 1
            print(f"转换失败：{path.name}：{exc}", file=sys.stderr)
        else:
            print(f"转换完成：{destination}")
    print(f"批量转换结束：成功 {len(files) - failed} 个，失败 {failed} 个")
    return failed


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("input", type=Path, help="原始 JSON 文件或输入目录")
    parser.add_argument("output", type=Path, nargs="?", help="输出 JSON 文件或目录")
    parser.add_argument(
        "--dir", action="store_true", dest="directory_mode",
        help="批量转换目录中的 JSON 文件，保留文件名；必须指定输出目录",
    )
    parser.add_argument(
        "--expires-in", type=int, default=2419200,
        help="expires_in 的值，单位秒，默认 2419200（28 天）",
    )
    args = parser.parse_args()
    if args.expires_in < 0:
        parser.error("--expires-in 不能小于 0")
    if args.directory_mode and args.output is None:
        parser.error("--dir 模式必须指定输入目录和输出目录")

    source = args.input
    try:
        if args.directory_mode:
            failed = convert_directory(source, args.output, args.expires_in)
            raise SystemExit(1 if failed else 0)
        output = args.output or source.with_name(source.stem + ".converted.json")
        convert_file(source, output, args.expires_in)
    except json.JSONDecodeError as exc:
        parser.exit(1, f"转换失败：JSON 格式错误，第 {exc.lineno} 行第 {exc.colno} 列\n")
    except (OSError, UnicodeError, ValueError) as exc:
        parser.exit(1, f"转换失败：{exc}\n")
    print(f"转换完成：{output}")


if __name__ == "__main__":
    main()
