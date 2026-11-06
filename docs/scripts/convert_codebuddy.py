#!/usr/bin/env python3
"""将 CodeBuddy account/auth JSON 转换为指定的 token JSON（仅使用标准库）。

用法：
    python convert_codebuddy.py input.json output.json
    python convert_codebuddy.py input.json
    python convert_codebuddy.py input.json output.json --expires-in 2419200
"""

import argparse
import json
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


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("input", type=Path, help="原始 JSON 文件")
    parser.add_argument("output", type=Path, nargs="?", help="输出 JSON 文件")
    parser.add_argument(
        "--expires-in", type=int, default=2419200,
        help="expires_in 的值，单位秒，默认 2419200（28 天）",
    )
    args = parser.parse_args()
    if args.expires_in < 0:
        parser.error("--expires-in 不能小于 0")

    source = args.input
    output = args.output or source.with_name(source.stem + ".converted.json")
    if source.resolve() == output.resolve():
        parser.error("输入和输出不能是同一个文件")

    try:
        # 兼容 Windows 编辑器生成的 UTF-8 BOM。
        data = json.loads(source.read_text(encoding="utf-8-sig"))
        result = convert(data, expires_in=args.expires_in)
        output.write_text(
            json.dumps(result, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
    except json.JSONDecodeError as exc:
        parser.exit(1, f"转换失败：JSON 格式错误，第 {exc.lineno} 行第 {exc.colno} 列\n")
    except (OSError, UnicodeError, ValueError) as exc:
        parser.exit(1, f"转换失败：{exc}\n")
    print(f"转换完成：{output}")


if __name__ == "__main__":
    main()
