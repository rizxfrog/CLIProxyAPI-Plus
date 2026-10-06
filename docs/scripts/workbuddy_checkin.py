#!/usr/bin/env python3
"""
WorkBuddy 每日签到脚本

调用腾讯 WorkBuddy/CodeBuddy 桌面端的「每日签到」(Buddy 加油站) 后端接口。

接口（从 app.asar 逆向提取）:
  - 查询状态:  POST {endpoint}/v2/billing/meter/checkin-activity-status
  - 执行签到:  POST {endpoint}/v2/billing/meter/daily-checkin
  - 刷新凭证:  POST {endpoint}/v2/plugin/auth/token/refresh   (仅 --refresh-token)

鉴权:
  - Authorization: Bearer <accessToken>
  - X-User-Id: <uid>
  - X-Enterprise-Id / X-Tenant-Id: 企业账号时存在
  - X-Domain: 域账号时存在

登录态来源（自动读取，也可手动传入）:
  Windows:  %LOCALAPPDATA%\\CodeBuddyExtension\\Data\\Public\\auth\\<id>.info
  macOS:    ~/Library/Application Support/CodeBuddyExtension/Data/Public/auth/<id>.info
  Linux:    ~/.local/share/CodeBuddyExtension/Data/Public/auth/<id>.info

刷新模式（--refresh-token）:
  先用凭证里的 refresh_token 调 /v2/plugin/auth/token/refresh 轮换 accessToken
  （头部 X-Refresh-Token + X-Auth-Refresh-Source: plugin，与 Go 端
  internal/auth/codebuddycn/codebuddycn.go 的 Client.Refresh 对齐），把新的
  accessToken/refreshToken/expired 原子写回原凭证文件，再用新 token 查询并签到。
  刷新失败时不发后续请求，直接按失败计；批量模式下继续处理其他凭证。
  该模式需要凭证文件（--auth-file/--auth-dir），不能与 --token/--uid 手动模式同用。
  注意：脚本默认不刷新，凭证过期时需先通过代理或桌面端登录更新。

用法:
  uv run workbuddy_checkin.py                     # 查询状态 + 签到
  uv run workbuddy_checkin.py status              # 仅查询状态
  uv run workbuddy_checkin.py --token <TOKEN> --uid <UID>   # 手动指定
  uv run workbuddy_checkin.py --refresh-token     # 刷新凭证后再查询 + 签到
  uv run workbuddy_checkin.py --auth-dir <DIR>    # 对目录下所有 .json 凭证文件签到
  uv run workbuddy_checkin.py --auth-dir <DIR> --refresh-token  # 批量刷新凭证并签到
  uv run workbuddy_checkin.py --auth-dir <DIR> --prefix codebuddy  # 只处理 codebuddy*.json
  uv run workbuddy_checkin.py --auth-file <AUTH_FILE> # 对单个文件执行签到
"""

from __future__ import annotations

import argparse
import base64
import json
import os
import sys
import time
from pathlib import Path
from urllib.parse import urlsplit

# 尽量零依赖：优先用 requests，否则退回标准库 urllib
try:
    import requests  # type: ignore
except ImportError:
    requests = None

# ---------------------------------------------------------------- 配置
DEFAULT_ENDPOINT = "https://copilot.tencent.com"
STAGING_ENDPOINT = "https://staging.codebuddy.cn"

# 凭证刷新（--refresh-token）：与 Go 端 internal/auth/codebuddycn 的
# Client.Refresh 使用同一个插件级刷新端点与头部。
REFRESH_PATH = "/v2/plugin/auth/token/refresh"
REFRESH_SOURCE = "plugin"
UA_VERSION = "5.3.14"

AUTH_ID = "auth"  # authenticationId，可用 authentication.id 覆盖
AUTH_REL_PATH = ("CodeBuddyExtension", "Data", "Public", "auth")

# 可能的登录态文件路径（按平台）
def _auth_candidates() -> list[Path]:
    home = Path.home()
    env_override = os.environ.get("WORKBUDDY_AUTH_FILE") or os.environ.get("CODEBUDDY_AUTH_FILE")
    if env_override:
        return [Path(env_override)]

    bases = []
    if sys.platform == "win32":
        local = os.environ.get("LOCALAPPDATA")
        if local:
            bases.append(Path(local))
        bases.append(home / "AppData" / "Local")
    elif sys.platform == "darwin":
        bases.append(home / "Library" / "Application Support")
    else:
        bases.append(home / ".local" / "share")
        bases.append(home / ".config")

    out = []
    for base in bases:
        d = base.joinpath(*AUTH_REL_PATH)
        out.append(d / f"{AUTH_ID}.info")
        # 也尝试目录下任意 .info 文件
        out.extend(sorted(d.glob("*.info")) if d.exists() else [])
    return out


# ---------------------------------------------------------------- 登录态
def _decode_jwt_sub(token: str) -> str | None:
    """从 JWT 的 payload 解出 sub（作为用户 uid）。失败返回 None。"""
    try:
        parts = token.split(".")
        if len(parts) < 2:
            return None
        pad = "=" * (-len(parts[1]) % 4)
        payload = json.loads(base64.urlsafe_b64decode(parts[1] + pad))
        return payload.get("sub")
    except Exception:
        return None


def load_session(auth_file: Path | None) -> dict:
    """读取登录态，返回含 token/uid/enterpriseId/domain/endpoint 的 dict。

    支持两种格式：
      1. 桌面端 auth.info：{ auth: { accessToken, domain }, account: { uid, enterpriseId } }
      2. CLI 凭证（codebuddy-cn 等）：{ access_token, base_url, refresh_token, ... }
         此时 uid 从 JWT 的 sub 自动解出，endpoint 取自 base_url。

    额外的 refresh_token / _layout / _path 供 --refresh-token 模式轮换并写回凭证。
    """
    path = auth_file or next((p for p in _auth_candidates() if p.is_file()), None)
    if path is None:
        raise SystemExit(
            "未找到登录态文件。请先用 WorkBuddy 登录，或通过 "
            "--auth-file / --token+--uid 指定。"
        )
    data = json.loads(path.read_text(encoding="utf-8"))

    # 格式 2：CLI 凭证
    if "access_token" in data:
        token = data.get("access_token")
        uid = _decode_jwt_sub(token)
        if not token or not uid:
            raise SystemExit(f"凭证文件 {path} 中 access_token 无效或无法解出 uid。")
        endpoint = (data.get("base_url") or "").rstrip("/")
        # base_url 常带 /v2，去掉它以配合脚本内部的 /v2 拼接
        if endpoint.endswith("/v2"):
            endpoint = endpoint[: -len("/v2")]
        print(f"[i] 登录态来源: {path} (CLI 凭证 type={data.get('type')})")
        return {
            "token": token,
            "uid": uid,
            "enterpriseId": data.get("enterprise_id") or data.get("enterpriseId"),
            "domain": data.get("domain"),
            "endpoint": endpoint or None,
            "refresh_token": (data.get("refresh_token") or data.get("refreshToken") or "").strip(),
            "_layout": "cli",
            "_path": str(path),
        }

    # 格式 1：桌面端 auth.info
    auth = data.get("auth", {}) or {}
    account = data.get("account", {}) or {}
    token = auth.get("accessToken")
    uid = account.get("uid")
    if not token or not uid:
        raise SystemExit(f"登录态文件 {path} 中缺少 accessToken/uid。")
    print(f"[i] 登录态来源: {path} (桌面端 auth.info)")
    return {
        "token": token,
        "uid": uid,
        "enterpriseId": account.get("enterpriseId"),
        "domain": auth.get("domain"),
        "endpoint": None,
        "refresh_token": (auth.get("refreshToken") or auth.get("refresh_token") or "").strip(),
        "_layout": "desktop",
        "_path": str(path),
    }


# ---------------------------------------------------------------- HTTP
def _http_json(url: str, headers: dict, body: dict, timeout: int) -> dict:
    if requests is not None:
        r = requests.post(url, headers=headers, json=body, timeout=timeout)
        status = r.status_code
        try:
            payload = r.json()
        except Exception:
            payload = {"raw": r.text}
    else:
        import urllib.error
        import urllib.request

        req = urllib.request.Request(
            url,
            data=json.dumps(body).encode("utf-8"),
            headers={**headers, "Content-Type": "application/json"},
            method="POST",
        )
        try:
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                status = resp.status
                payload = json.loads(resp.read().decode("utf-8"))
        except urllib.error.HTTPError as e:
            status = e.code
            try:
                payload = json.loads(e.read().decode("utf-8"))
            except Exception:
                payload = {"raw": e.read().decode("utf-8", "replace")}
    return {"status": status, "payload": payload}


def build_headers(session: dict, ua_version: str = UA_VERSION) -> dict:
    h = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": f"Bearer {session['token']}",
        "X-User-Id": str(session["uid"]),
        "User-Agent": f"WorkBuddy/{ua_version}",
    }
    if session.get("enterpriseId"):
        h["X-Enterprise-Id"] = str(session["enterpriseId"])
        h["X-Tenant-Id"] = str(session["enterpriseId"])
    if session.get("domain"):
        h["X-Domain"] = str(session["domain"])
    return h


# ---------------------------------------------------------------- 凭证刷新
def _refresh_domain(endpoint: str, session: dict) -> str:
    """X-Domain：优先凭证里的域，否则用 endpoint 的 host（与 Go 端 domain 默认值一致）。"""
    domain = str(session.get("domain") or "").strip()
    if domain:
        return domain
    return urlsplit(endpoint).netloc or "copilot.tencent.com"


def refresh_credential(endpoint: str, session: dict, timeout: int) -> dict:
    """用 refresh_token 换新 token 对，返回 {token, refresh_token, expires_in, expired}。

    头与 Go 端 internal/auth/codebuddycn.Client.Refresh 对齐：插件级刷新端点，
    X-Refresh-Token + X-Auth-Refresh-Source: plugin，body 为 {}。
    """
    refresh_token = (session.get("refresh_token") or "").strip()
    if not refresh_token:
        raise RuntimeError("凭证缺少 refresh_token，无法刷新（--refresh-token 需要可轮换的凭证文件）")

    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "User-Agent": f"WorkBuddy/{UA_VERSION}",
        "X-Requested-With": "XMLHttpRequest",
        "X-Domain": _refresh_domain(endpoint, session),
        "X-No-Authorization": "true",
        "X-No-User-Id": "true",
        "X-Product": "SaaS",
        "X-Refresh-Token": refresh_token,
        "X-Auth-Refresh-Source": REFRESH_SOURCE,
    }
    r = _http_json(f"{endpoint}{REFRESH_PATH}", headers, {}, timeout)
    payload = r.get("payload")
    if r["status"] != 200:
        detail = payload.get("msg") if isinstance(payload, dict) else None
        raise RuntimeError(f"刷新失败 HTTP {r['status']}: {detail or ''}".rstrip())
    code = payload.get("code") if isinstance(payload, dict) else None
    if code != 0:
        msg = payload.get("msg") if isinstance(payload, dict) else None
        raise RuntimeError(f"刷新被拒绝 code={code}: {msg or ''}".rstrip())
    data = payload.get("data") if isinstance(payload.get("data"), dict) else {}
    access = (data.get("accessToken") or "").strip()
    if not access:
        raise RuntimeError("刷新响应缺少 accessToken")
    expires_in = data.get("expiresIn") or 0
    expired = ""
    try:
        if int(expires_in) > 0:
            expired = time.strftime(
                "%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + int(expires_in))
            )
    except (TypeError, ValueError):
        expires_in = 0
    return {
        "token": access,
        "refresh_token": (data.get("refreshToken") or refresh_token).strip(),
        "expires_in": int(expires_in or 0),
        "expired": expired,
    }


def save_session(session: dict, refreshed: dict) -> None:
    """把轮换后的凭证原子写回原文件，保留其余字段（对齐 Go 端 filestore 行为）。"""
    raw_path = session.get("_path")
    if not raw_path:
        raise RuntimeError("手动 --token 模式没有凭证文件，无法写回刷新结果")
    path = Path(raw_path)
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as e:
        raise RuntimeError(f"回写前读取凭证失败: {path}: {e}") from None

    if session.get("_layout") == "desktop":
        # auth.info is the desktop client's own file; only touch fields inside
        # "auth" so no foreign keys leak into its schema.
        target = data.setdefault("auth", {})
        target["accessToken"] = refreshed["token"]
        if refreshed["refresh_token"]:
            target["refreshToken"] = refreshed["refresh_token"]
    else:
        # CLIProxyAPI auth file: mirror the metadata keys the Go filestore writes.
        data["access_token"] = refreshed["token"]
        if refreshed["refresh_token"]:
            data["refresh_token"] = refreshed["refresh_token"]
        if refreshed["expires_in"] > 0:
            data["expires_in"] = refreshed["expires_in"]
        if refreshed["expired"]:
            data["expired"] = refreshed["expired"]
        data["timestamp"] = int(time.time() * 1000)

    tmp = path.with_name(path.name + ".tmp")
    tmp.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    os.replace(tmp, path)


def ensure_fresh_credential(endpoint: str, session: dict, timeout: int) -> dict:
    """--refresh-token 模式：刷新凭证、写回文件，并更新内存中的会话。"""
    print("[i] 刷新登录凭证...")
    refreshed = refresh_credential(endpoint, session, timeout)
    session = dict(session)
    session["token"] = refreshed["token"]
    session["refresh_token"] = refreshed["refresh_token"]
    uid = _decode_jwt_sub(refreshed["token"])
    if uid:
        session["uid"] = uid
    save_session(session, refreshed)
    print(f"[✓] 凭证已刷新并写回 {session['_path']}"
          + (f"（expired={refreshed['expired']}）" if refreshed["expired"] else ""))
    return session


# ---------------------------------------------------------------- 业务
def checkin_status(endpoint: str, session: dict, timeout: int) -> dict:
    url = f"{endpoint}/v2/billing/meter/checkin-activity-status"
    return _http_json(url, build_headers(session), {}, timeout)


def claim_daily_checkin(endpoint: str, session: dict, timeout: int) -> dict:
    url = f"{endpoint}/v2/billing/meter/daily-checkin"
    return _http_json(url, build_headers(session), {}, timeout)


def _pretty(data: dict) -> None:
    print(json.dumps(data, ensure_ascii=False, indent=2))


def _run_one(args: argparse.Namespace, session: dict, endpoint: str) -> int:
    """对单个登录态执行 status/checkin（--refresh-token 时先轮换凭证）。"""
    print(f"[i] endpoint: {endpoint}")

    if args.refresh_token:
        # 刷新失败时不再发后续请求，避免用已知过期的凭证误判。
        try:
            session = ensure_fresh_credential(endpoint, session, args.timeout)
        except Exception as e:
            print(f"[✗] 凭证刷新失败: {e}")
            return 1

    if args.action == "status":
        r = checkin_status(endpoint, session, args.timeout)
        _pretty(r)
        return 0 if r["status"] == 200 else 1

    # checkin: 先查状态再签到
    print("[i] 查询签到状态...")
    st = checkin_status(endpoint, session, args.timeout)
    _pretty(st)
    data = st.get("payload", {}).get("data") if isinstance(st.get("payload"), dict) else None
    already = (
        data.get("today_checked_in")
        if isinstance(data, dict)
        else False
    )
    if already:
        print("\n[✓] 今日已签到，无需重复。")
        return 0

    print("\n[i] 执行签到...")
    cl = claim_daily_checkin(endpoint, session, args.timeout)
    _pretty(cl)
    if cl["status"] == 200:
        payload = cl.get("payload", {})
        code = payload.get("code") if isinstance(payload, dict) else None
        if code == 0:
            print("\n[✓] 签到成功！")
            return 0
    print("\n[✗] 签到未成功（见上方响应）。")
    return 1


def _auth_dir_files(auth_dir: Path, prefix: str | None = None) -> list[Path]:
    """返回目录下所有 .json 凭证文件（排序后），可按文件名前缀过滤。"""
    if not auth_dir.is_dir():
        raise SystemExit(f"--auth-dir 指定的目录不存在或不是目录: {auth_dir}")
    files = sorted(auth_dir.glob("*.json"))
    if prefix:
        files = [f for f in files if f.name.startswith(prefix)]
    if not files:
        suffix = f"（前缀 {prefix!r}）" if prefix else ""
        raise SystemExit(f"--auth-dir 指定的目录下没有匹配的 .json 文件{suffix}: {auth_dir}")
    return files


def main(argv: list[str] | None = None) -> int:
    ap = argparse.ArgumentParser(description="WorkBuddy 每日签到")
    ap.add_argument("action", nargs="?", default="checkin", choices=["status", "checkin"])
    ap.add_argument("--endpoint", default=os.environ.get("WORKBUDDY_ENDPOINT", DEFAULT_ENDPOINT))
    ap.add_argument("--staging", action="store_true", help="使用 staging 环境")
    ap.add_argument("--auth-file", type=Path, default=None, help="手动指定登录态文件路径")
    ap.add_argument(
        "--auth-dir",
        type=Path,
        default=None,
        help="指定 auth 目录，对其下所有 .json 文件依次执行签到",
    )
    ap.add_argument(
        "--prefix",
        default=None,
        help="配合 --auth-dir 使用，只处理文件名以该前缀开头的 .json 文件",
    )
    ap.add_argument("--token", default=None)
    ap.add_argument("--uid", default=None)
    ap.add_argument("--enterprise-id", default=None)
    ap.add_argument("--domain", default=None)
    ap.add_argument(
        "--refresh-token",
        action="store_true",
        help="签到前先用凭证里的 refresh_token 轮换 token 并写回凭证文件",
    )
    ap.add_argument("--timeout", type=int, default=15)
    args = ap.parse_args(argv)

    if args.staging:
        args.endpoint = STAGING_ENDPOINT
    endpoint = args.endpoint.rstrip("/")

    if args.prefix and not args.auth_dir:
        raise SystemExit("--prefix 需要配合 --auth-dir 使用")

    if args.refresh_token and args.token:
        raise SystemExit("--refresh-token 需要凭证文件（--auth-file/--auth-dir），不能与 --token 同用")

    if args.token and args.uid:
        session = {
            "token": args.token,
            "uid": args.uid,
            "enterpriseId": args.enterprise_id,
            "domain": args.domain,
            "endpoint": None,
        }
        return _run_one(args, session, endpoint)

    if args.auth_dir:
        total = 0
        for path in _auth_dir_files(args.auth_dir, args.prefix):
            print("\n" + "=" * 60)
            print(f"[i] 处理凭证文件: {path}")
            print("=" * 60)
            try:
                session = load_session(path)
            except SystemExit as e:
                print(f"[✗] 跳过 {path}: {e}")
                total += 1
                continue
            file_endpoint = endpoint
            # 凭证文件里若带 endpoint/base_url，优先使用（未显式传 --endpoint 时）
            if session.get("endpoint") and args.endpoint == DEFAULT_ENDPOINT:
                file_endpoint = session["endpoint"].rstrip("/")
            total += _run_one(args, session, file_endpoint)
        return total

    session = load_session(args.auth_file)

    # 凭证文件里若带 endpoint/base_url，优先使用（未显式传 --endpoint 时）
    if session.get("endpoint") and args.endpoint == DEFAULT_ENDPOINT:
        endpoint = session["endpoint"].rstrip("/")

    return _run_one(args, session, endpoint)


if __name__ == "__main__":
    raise SystemExit(main())
