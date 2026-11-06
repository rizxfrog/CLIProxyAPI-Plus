#!/usr/bin/env python3
"""
LobsterAI（网易有道龙虾）每日签到脚本

调用 LobsterAI 桌面端的「每日积分礼」(客户端活动) 后端接口。

接口（从 resources/app.asar → dist-electron/main/libs/activity/activityClient.js
与 shared/activity/constants.js 逆向提取）:
  - 查询槽位:  GET  {backend}/api/client-activities/slot?placement=&clientVersion=&containerApiVersion=&platform=
  - 查询上下文: GET  {backend}/api/client-activities/{activityCode}/context?configRevision=
  - 执行动作:  POST {backend}/api/client-activities/{activityCode}/actions/{actionId}
                     body: {"configRevision": N, "idempotencyKey": "...", "payload": {}}
  - 刷新凭证:  POST {backend}/api/auth/refresh   (仅 --refresh-token)

鉴权:
  - Authorization: Bearer <accessToken>
  - X-LobsterAI-Client-Capabilities: kimi-k3-agentic-v1,thinking-level-control-v1
  - X-LobsterAI-Client-Version: <clientVersion>

关键实测结论（2026-10-15，免费账号）:
  - platform 必须是 win32 / darwin / linux；传 "cli" 或留空，服务端返回
    slotState="empty"（活动被隐藏），拿不到活动与签到入口。
  - placement=desktop_sidebar + containerApiVersion=2 是每日签到所在的槽位。
  - 上下文 state.claimedToday 表示今日是否已领；actions 含 "check_in" 才可签到。
  - 错误码: 51100 未找到 / 51101 未开始 / 51102 未登录 / 51103 动作非法 /
    51104 今日已领 / 51105 配置非法 / 51106 版本不匹配。
  - activityCode 形如 daily-check-in-evergreen-prod-20260814（常驻活动）。

登录态来源（自动读取，也可手动传入）:
  命令行:  --auth-file <FILE> / --auth-dir <DIR> [--prefix lobsterai-server]
  环境变量: LOBSTERAI_SERVER_AUTH_FILE / LOBSTERAI_AUTH_FILE
  默认:    ./data/auth_files/lobsterai-server*.json（从仓库根运行时）

用法:
  uv run lobsterai_server_checkin.py                  # 查询状态 + 签到
  uv run lobsterai_server_checkin.py status            # 仅查询状态
  uv run lobsterai_server_checkin.py --token <TOKEN>   # 手动指定凭证
  uv run lobsterai_server_checkin.py --refresh-token   # 先刷新凭证再查询 + 签到
  uv run lobsterai_server_checkin.py --auth-file <F>   # 对单个凭证文件签到
  uv run lobsterai_server_checkin.py --auth-dir <DIR>  # 对目录下所有 .json 凭证签到
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import time
from pathlib import Path
from urllib.parse import urlencode

# 尽量零依赖：优先用 requests，否则退回标准库 urllib
try:
    import requests  # type: ignore
except ImportError:
    requests = None

# ---------------------------------------------------------------- 配置
DEFAULT_BACKEND = "https://lobsterai-server.youdao.com"
TEST_BACKEND = "https://lobsterai-server.inner.youdao.com"

DEFAULT_CLIENT_VERSION = "2026.10.13"
CLIENT_CAPABILITIES = "kimi-k3-agentic-v1,thinking-level-control-v1"

# 每日签到所在槽位（shared/activity/constants.js: ActivityContainerApiVersion.NativeDailyCheckInV1 = 2）
DEFAULT_PLACEMENT = "desktop_sidebar"
DEFAULT_CONTAINER_API_VERSION = 2
# 备用槽位：一次性积分回馈（NativeStartupCreditV1 = 3）
STARTUP_PLACEMENT = "desktop_startup_modal"
STARTUP_CONTAINER_API_VERSION = 3

# 校验规则（与客户端 ACTIVITY_CODE_PATTERN / IDEMPOTENCY_KEY_PATTERN 一致）
ACTIVITY_CODE_PATTERN = r"^[a-z0-9][a-z0-9_-]{2,63}$"
IDEMPOTENCY_KEY_PATTERN = r"^[A-Za-z0-9._:-]{1,64}$"

REFRESH_PATH = "/api/auth/refresh"


def _platform() -> str:
    """客户端上报的平台标识。'cli' 会被服务端当作无活动，必须用真实平台值。"""
    if sys.platform == "win32":
        return "win32"
    if sys.platform == "darwin":
        return "darwin"
    return "linux"


def _default_auth_globs() -> list[Path]:
    """未显式指定凭证时的候选路径。"""
    override = os.environ.get("LOBSTERAI_SERVER_AUTH_FILE") or os.environ.get("LOBSTERAI_AUTH_FILE")
    if override:
        return [Path(override)]
    candidates: list[Path] = []
    for root in (Path.cwd(), *Path.cwd().parents):
        d = root / "data" / "auth_files"
        if d.is_dir():
            candidates.extend(sorted(d.glob("lobsterai-server*.json")))
            candidates.extend(sorted(d.glob("lobsterai*.json")))
            break
    return candidates


# ---------------------------------------------------------------- HTTP
def _http_json(url: str, headers: dict, body: dict | None, timeout: int, method: str = "POST") -> dict:
    """发一个 JSON 请求，返回 {status, payload}。payload 解析失败时含 raw。"""
    if requests is not None:
        r = requests.request(method, url, headers=headers, json=body, timeout=timeout)
        status = r.status_code
        try:
            payload = r.json()
        except Exception:
            payload = {"raw": r.text}
        return {"status": status, "payload": payload}

    import urllib.error
    import urllib.request

    data = None
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers = {**headers, "Content-Type": "application/json"}
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return {"status": resp.status, "payload": json.loads(resp.read().decode("utf-8"))}
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", "replace")
        try:
            payload = json.loads(raw)
        except Exception:
            payload = {"raw": raw}
        return {"status": e.code, "payload": payload}


def build_headers(session: dict) -> dict:
    headers = {
        "Accept": "application/json",
        "X-LobsterAI-Client-Capabilities": CLIENT_CAPABILITIES,
        "X-LobsterAI-Client-Version": session.get("client_version") or DEFAULT_CLIENT_VERSION,
        "User-Agent": "CLIProxyAPI/lobsterai-server",
    }
    # The refresh endpoint is unauthenticated (the refresh token is in the body),
    # so a tokenless session is legitimate there.
    token = (session.get("token") or "").strip()
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


# ---------------------------------------------------------------- 登录态
def load_session(auth_file: Path | None, token: str | None, backend: str | None) -> dict:
    """读取登录态，返回含 token/backend/client_version 的 dict。

    CLIProxyAPI 凭证文件形如:
      {"access_token": "...", "refresh_token": "...", "backend_url": "...",
       "base_url": ".../api/proxy/v1", "client_version": "...", "type": "lobsterai-server"}
    """
    if token:
        return {
            "token": token,
            "backend": (backend or DEFAULT_BACKEND).rstrip("/"),
            "client_version": DEFAULT_CLIENT_VERSION,
            "refresh_token": "",
            "_path": None,
        }

    path = auth_file
    if path is None:
        candidates = _default_auth_globs()
        path = next((p for p in candidates if p.is_file()), None)
    if path is None:
        raise SystemExit(
            "未找到登录态文件。请用 --auth-file / --auth-dir 指定，或用 --token 手动传入，"
            "或设置 LOBSTERAI_SERVER_AUTH_FILE。"
        )
    try:
        data = json.loads(Path(path).read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as e:
        raise SystemExit(f"读取凭证文件失败 {path}: {e}") from None

    access = (data.get("access_token") or "").strip()
    if not access:
        raise SystemExit(f"凭证文件 {path} 缺少 access_token。")
    resolved_backend = (backend or data.get("backend_url") or "").rstrip("/")
    if not resolved_backend:
        # 从 base_url（.../api/proxy/v1）反推后端 origin
        base = (data.get("base_url") or "").strip()
        suffix = "/api/proxy/v1"
        resolved_backend = base[: -len(suffix)].rstrip("/") if base.endswith(suffix) else DEFAULT_BACKEND
    print(f"[i] 登录态来源: {path} (type={data.get('type')})")
    return {
        "token": access,
        "backend": resolved_backend or DEFAULT_BACKEND,
        "client_version": (data.get("client_version") or "").strip() or DEFAULT_CLIENT_VERSION,
        "refresh_token": (data.get("refresh_token") or "").strip(),
        "_path": str(path),
    }


# ---------------------------------------------------------------- 凭证刷新
def refresh_credential(session: dict, timeout: int) -> dict:
    """用 refresh_token 换新 token 对。

    body 与客户端一致：{refreshToken, firstKeyfrom, latestKeyfrom, version}。
    响应 {code:0,data:{accessToken,refreshToken?}}；未返回 refreshToken 时保留旧的。
    """
    refresh_token = (session.get("refresh_token") or "").strip()
    if not refresh_token:
        raise RuntimeError("凭证缺少 refresh_token，无法刷新（--refresh-token 需要可轮换的凭证文件）")

    body = {
        "refreshToken": refresh_token,
        "firstKeyfrom": "official",
        "latestKeyfrom": "official",
        "version": session.get("client_version") or DEFAULT_CLIENT_VERSION,
    }
    r = _http_json(f"{session['backend']}{REFRESH_PATH}", build_headers(session), body, timeout)
    payload = r.get("payload")
    if r["status"] != 200:
        raise RuntimeError(f"刷新失败 HTTP {r['status']}")
    if not isinstance(payload, dict) or payload.get("code") != 0:
        code = payload.get("code") if isinstance(payload, dict) else None
        msg = payload.get("message") if isinstance(payload, dict) else None
        raise RuntimeError(f"刷新被拒绝 code={code}: {msg or ''}".rstrip())
    data = payload.get("data") if isinstance(payload.get("data"), dict) else {}
    access = (data.get("accessToken") or "").strip()
    if not access:
        raise RuntimeError("刷新响应缺少 accessToken")
    expires_in = int(data.get("expiresIn") or 0)
    expired = ""
    if expires_in > 0:
        expired = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + expires_in))
    return {
        "token": access,
        "refresh_token": (data.get("refreshToken") or refresh_token).strip(),
        "expires_in": expires_in,
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


def ensure_fresh_credential(session: dict, timeout: int) -> dict:
    refreshed = refresh_credential(session, timeout)
    session = dict(session)
    session["token"] = refreshed["token"]
    session["refresh_token"] = refreshed["refresh_token"]
    save_session(session, refreshed)
    print(
        f"[✓] 凭证已刷新并写回 {session['_path']}"
        + (f"（expired={refreshed['expired']}）" if refreshed["expired"] else "")
    )
    return session


# ---------------------------------------------------------------- 业务
def fetch_slot(session: dict, placement: str, container_api_version: int, timeout: int) -> dict:
    query = urlencode(
        {
            "placement": placement,
            "clientVersion": session.get("client_version") or DEFAULT_CLIENT_VERSION,
            "containerApiVersion": container_api_version,
            "platform": _platform(),
        }
    )
    url = f"{session['backend']}/api/client-activities/slot?{query}"
    return _http_json(url, build_headers(session), None, timeout, method="GET")


def fetch_context(session: dict, activity_code: str, config_revision: int, timeout: int) -> dict:
    query = urlencode({"configRevision": config_revision})
    url = (
        f"{session['backend']}/api/client-activities/{activity_code}/context?{query}"
    )
    return _http_json(url, build_headers(session), None, timeout, method="GET")


def execute_action(
    session: dict, activity_code: str, action_id: str, config_revision: int, timeout: int
) -> dict:
    # 幂等键：唯一、且匹配 ^[A-Za-z0-9._:-]{1,64}$
    idempotency_key = f"cli-{int(time.time() * 1000)}-{os.urandom(8).hex()}"
    body = {
        "configRevision": config_revision,
        "idempotencyKey": idempotency_key,
        "payload": {},
    }
    url = f"{session['backend']}/api/client-activities/{activity_code}/actions/{action_id}"
    return _http_json(url, build_headers(session), body, timeout)


def _pretty(data: dict) -> None:
    print(json.dumps(data, ensure_ascii=False, indent=2))


def _payload_data(result: dict) -> dict:
    payload = result.get("payload")
    if isinstance(payload, dict) and isinstance(payload.get("data"), dict):
        return payload["data"]
    return {}


def _describe_activity(slot_data: dict) -> tuple[str | None, int]:
    activity = slot_data.get("activity") if isinstance(slot_data, dict) else None
    if not isinstance(activity, dict):
        return None, 0
    code = activity.get("activityCode")
    revision = activity.get("configRevision")
    return (code if isinstance(code, str) else None, revision if isinstance(revision, int) else 0)


def _run_one(args: argparse.Namespace, session: dict) -> int:
    """对单个登录态执行 status/checkin（--refresh-token 时先轮换凭证）。"""
    print(f"[i] backend: {session['backend']}  platform: {_platform()}")

    if args.refresh_token:
        # 刷新失败时不发后续请求，避免用已知过期的凭证误判。
        try:
            session = ensure_fresh_credential(session, args.timeout)
        except Exception as e:
            print(f"[✗] 凭证刷新失败: {e}")
            return 1

    slots = [(DEFAULT_PLACEMENT, DEFAULT_CONTAINER_API_VERSION)]
    if args.include_startup:
        slots.append((STARTUP_PLACEMENT, STARTUP_CONTAINER_API_VERSION))

    chosen: tuple[str, int, dict] | None = None
    for placement, api_version in slots:
        print(f"\n[i] 查询槽位 placement={placement} containerApiVersion={api_version} ...")
        slot = fetch_slot(session, placement, api_version, args.timeout)
        _pretty(slot)
        if slot["status"] != 200:
            continue
        data = _payload_data(slot)
        code, revision = _describe_activity(data)
        if data.get("slotState") == "available" and code and revision > 0:
            chosen = (code, revision, data)
            break

    if chosen is None:
        print("\n[✓] 当前没有可用的活动槽位（slotState=empty），无需处理。")
        return 0

    activity_code, config_revision, descriptor = chosen
    activity = descriptor.get("activity") if isinstance(descriptor.get("activity"), dict) else {}
    print(
        f"\n[i] 活动: {activity.get('activityType')} code={activity_code} "
        f"rev={config_revision} title={activity.get('cardTitle')}"
    )

    print("\n[i] 查询活动上下文...")
    ctx = fetch_context(session, activity_code, config_revision, args.timeout)
    _pretty(ctx)
    if ctx["status"] != 200:
        print("\n[✗] 读取活动上下文失败。")
        return 1
    ctx_data = _payload_data(ctx)
    state = ctx_data.get("state") if isinstance(ctx_data.get("state"), dict) else {}
    actions = ctx_data.get("actions") if isinstance(ctx_data.get("actions"), list) else []

    if args.action == "status":
        return 0

    if state.get("completed"):
        print("\n[✓] 本期活动已完成，无需处理。")
        return 0
    if state.get("claimedToday"):
        print(
            f"\n[✓] 今日已签到（累计 {state.get('claimedDays')} 天，"
            f"已领 {state.get('claimedCredits')} 积分）。"
        )
        return 0

    action_id = "check_in" if "check_in" in actions else (actions[0] if actions else None)
    if not action_id:
        print("\n[✗] 活动未暴露任何可执行动作。")
        return 1

    print(f"\n[i] 执行动作 {action_id}（每日奖励 {state.get('rewardCredits')} 积分）...")
    result = execute_action(session, activity_code, action_id, config_revision, args.timeout)
    _pretty(result)

    payload = result.get("payload") if isinstance(result.get("payload"), dict) else {}
    code = payload.get("code")
    # 51104：并发/重复点击时服务端已记账，视为成功。
    if result["status"] == 200 and code == 0:
        print("\n[✓] 签到成功！")
        return 0
    if code == 51104:
        print("\n[✓] 今日已签到（服务端返回 51104 AlreadyClaimed）。")
        return 0
    print(f"\n[✗] 签到未成功（HTTP {result['status']} code={code}）。")
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
    ap = argparse.ArgumentParser(description="LobsterAI 每日签到")
    ap.add_argument("action", nargs="?", default="checkin", choices=["status", "checkin"])
    ap.add_argument("--backend", default=os.environ.get("LOBSTERAI_SERVER_BACKEND", DEFAULT_BACKEND))
    ap.add_argument("--test", action="store_true", help="使用内网测试环境")
    ap.add_argument("--auth-file", type=Path, default=None, help="手动指定凭证文件路径")
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
    ap.add_argument("--token", default=None, help="直接指定 access_token（跳过凭证文件）")
    ap.add_argument(
        "--refresh-token",
        action="store_true",
        help="签到前先用凭证里的 refresh_token 轮换 token 并写回凭证文件",
    )
    ap.add_argument(
        "--include-startup",
        action="store_true",
        help="备用槽位 desktop_startup_modal 也一并尝试（一次性积分回馈）",
    )
    ap.add_argument("--timeout", type=int, default=15)
    args = ap.parse_args(argv)

    if args.test:
        args.backend = TEST_BACKEND
    args.backend = (args.backend or DEFAULT_BACKEND).rstrip("/")

    if args.prefix and not args.auth_dir:
        raise SystemExit("--prefix 需要配合 --auth-dir 使用")
    if args.refresh_token and args.token:
        raise SystemExit("--refresh-token 需要凭证文件（--auth-file/--auth-dir），不能与 --token 同用")

    if args.token:
        session = load_session(None, args.token, args.backend)
        return _run_one(args, session)

    if args.auth_dir:
        total = 0
        for path in _auth_dir_files(args.auth_dir, args.prefix):
            print("\n" + "=" * 60)
            print(f"[i] 处理凭证文件: {path}")
            print("=" * 60)
            try:
                session = load_session(path, None, args.backend)
            except SystemExit as e:
                print(f"[✗] 跳过 {path}: {e}")
                total += 1
                continue
            total += _run_one(args, session)
        return total

    session = load_session(args.auth_file, None, args.backend)
    return _run_one(args, session)


if __name__ == "__main__":
    raise SystemExit(main())
