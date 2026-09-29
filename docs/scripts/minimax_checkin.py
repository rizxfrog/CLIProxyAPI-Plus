#!/usr/bin/env python3
"""MiniMax Code daily sign-in (签到) for CLIProxyAPI credentials.

Reversed from the MiniMax Code TUI client (MIT, commit 0f6ad52):
  - packages/tui/src/checkin/http-gateway.ts
  - packages/tui/src/runtime/public-gateway.ts
  - packages/shared/src/daily-signin.ts
See ../providers/minimax-integration.md for the provider integration notes.

Endpoints (signed public gateway):
  - Status: GET  {origin}/minimax-cloud/api/v1/signin/status
  - Claim:  POST {origin}/minimax-cloud/api/v1/signin/claim   (body "{}")
  - Identity (for user_id): GET {origin}/v1/api/user/info

Origins (prod):
  - en: https://agent.minimax.io        (CLIProxyAPI type "minimax")
  - cn: https://agent.minimaxi.com      (CLIProxyAPI type "minimax-cn")

Auth: Authorization: Bearer <access_token>. Requests also carry first-party
client attribution headers (yy, x-timestamp, x-signature); these are shared
wire constants, not credentials, and the Bearer token is the real authorization.

Usage:
  python3 minimax_checkin.py                              # status + check-in
  python3 minimax_checkin.py status                       # status only
  python3 minimax_checkin.py --auth-file <FILE>
  python3 minimax_checkin.py --auth-dir <DIR> --prefix minimax
  python3 minimax_checkin.py --token <TOKEN> [--region en|cn] [--user-id <ID>]
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

# Public gateway origins per region (prod). The CN *account* host is
# agent.minimaxi.com even though CN inference uses agent.minimax.cn.
ORIGINS = {
    "en": "https://agent.minimax.io",
    "cn": "https://agent.minimaxi.com",
}
ALLOWED_ORIGINS = frozenset(ORIGINS.values())

STATUS_PATH = "/minimax-cloud/api/v1/signin/status"
CLAIM_PATH = "/minimax-cloud/api/v1/signin/claim"
USER_INFO_PATH = "/v1/api/user/info"

# Wire-protocol constants copied verbatim from public-gateway.ts.
SIGNATURE_SALT = "I*7Cf%WZ#S&%1RlZJ&C2"
SIGNATURE_SUFFIX = "ooui"
VERSION_CODE = "22201"
APP_ID = "3001"
BIZ_ID = "3"
DEFAULT_APP_VERSION = "0.5.5"

# SigninDayStatus / SigninClaimResult from daily-signin.ts.
DAY_UPCOMING, DAY_CLAIMABLE, DAY_CLAIMED, DAY_DISABLED = 1, 2, 3, 4
CLAIM_CLAIMED, CLAIM_ALREADY_CLAIMED = 1, 2


class CheckinError(Exception):
    """A sanitized error safe to display without exposing credentials."""


class NoRedirect(urllib.request.HTTPRedirectHandler):
    """Never forward the bearer token through a redirect."""

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def md5(value: str) -> str:
    # MD5 is required by the client's request-signing wire format; it is an
    # integrity marker, not a security primitive, hence usedforsecurity=False.
    return hashlib.md5(value.encode("utf-8"), usedforsecurity=False).hexdigest()


# JavaScript encodeURIComponent leaves these characters unescaped.
_UNRESERVED = frozenset(
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.!~*'()"
)


def encode_uri_component(value: str) -> str:
    """Reproduce JavaScript's encodeURIComponent for signature compatibility."""
    out = []
    for byte in value.encode("utf-8"):
        char = chr(byte)
        if char in _UNRESERVED:
            out.append(char)
        else:
            out.append(f"%{byte:02X}")
    return "".join(out)


def timezone_offset_seconds() -> int:
    """Seconds east of UTC, matching new Date().getTimezoneOffset() * -60."""
    if time.timezone == 0:
        return 0
    is_dst = time.localtime().tm_isdst > 0
    offset = -(time.altzone if is_dst else time.timezone)
    return offset


def _query(region: str, real_user_id: str, now_ms: int, app_version: str, platform: str) -> list[tuple[str, str]]:
    language = "zh" if region == "cn" else "en"
    return [
        ("device_platform", "web"),
        ("biz_id", BIZ_ID),
        ("app_id", APP_ID),
        ("version_code", VERSION_CODE),
        ("is_desktop", "1"),
        ("desktop_version", app_version),
        ("unix", str(now_ms)),
        ("timezone_offset", str(timezone_offset_seconds())),
        ("sys_language", language),
        ("lang", language),
        ("device_id", "0"),
        ("os_name", platform),
        ("browser_name", "mcode"),
        ("user_id", real_user_id),
        ("client", "mcode"),
    ]


def _identity_query(region: str, real_user_id: str, now_ms: int, platform: str) -> list[tuple[str, str]]:
    language = "zh" if region == "cn" else "en"
    return [
        ("device_platform", "mcode"),
        ("biz_id", BIZ_ID),
        ("app_id", APP_ID),
        ("version_code", VERSION_CODE),
        ("unix", str(now_ms)),
        ("timezone_offset", str(timezone_offset_seconds())),
        ("sys_language", language),
        ("lang", language),
        ("device_id", "0"),
        ("os_name", platform),
        ("browser_name", "mcode"),
        ("user_id", real_user_id or "0"),
        ("client", "mcode"),
    ]


def build_request(
    origin: str,
    path: str,
    method: str,
    token: str,
    query: list[tuple[str, str]],
    body: str | None,
    now_ms: int,
) -> urllib.request.Request:
    """Build a signed public-gateway request exactly as the client does."""
    url = origin + path
    if query:
        # Preserve insertion order; the signature covers the exact path+query sent.
        url += "?" + urllib.parse.urlencode(query)
    parsed = urllib.parse.urlsplit(url)
    if f"{parsed.scheme}://{parsed.netloc}" not in ALLOWED_ORIGINS:
        raise CheckinError("Untrusted API URL")

    path_with_search = parsed.path
    if parsed.query:
        path_with_search += "?" + parsed.query
    signature_body = body if body is not None else ""
    yy_body = body if body is not None else "{}"
    second = now_ms // 1000
    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "User-Agent": "MiniMaxCode",
        "Authorization": f"Bearer {token}",
        "yy": md5(f"{encode_uri_component(path_with_search)}_{yy_body}{md5(str(now_ms))}{SIGNATURE_SUFFIX}"),
        "x-timestamp": str(second),
        "x-signature": md5(f"{second}{SIGNATURE_SALT}{signature_body}"),
    }
    # Scheme and authority are allowlisted above; redirects are disabled at open time.
    return urllib.request.Request(  # noqa: S310
        url,
        data=body.encode("utf-8") if body is not None else None,
        headers=headers,
        method=method,
    )


def request_json(request: urllib.request.Request, timeout: float = 20.0) -> dict:
    opener = urllib.request.build_opener(NoRedirect())
    try:
        with opener.open(request, timeout=timeout) as response:
            if not 200 <= response.status < 300:
                raise CheckinError(f"HTTP {response.status}")
            try:
                return json.load(response)
            except (ValueError, UnicodeError):
                raise CheckinError("Invalid JSON response") from None
    except urllib.error.HTTPError as error:
        status = error.code
        error.close()
        raise CheckinError(f"HTTP {status}") from None
    except CheckinError:
        raise
    except Exception:
        raise CheckinError("Network request failed; no automatic retry performed") from None


def _check_base_resp(payload: dict) -> None:
    base = payload.get("base_resp")
    if isinstance(base, dict):
        code = base.get("status_code")
        if isinstance(code, int) and code != 0:
            message = base.get("status_msg") or "request failed"
            raise CheckinError(f"Service rejected the request: {message}")


def _data(payload: dict):
    _check_base_resp(payload)
    data = payload.get("data")
    if data is None:
        raise CheckinError("Response data is missing")
    return data


def resolve_user_id(origin: str, token: str, region: str, transport, now_ms: int, app_version: str,
                    platform: str, timeout: float) -> str:
    """Resolve the account's realUserID via /v1/api/user/info."""
    request = build_request(
        origin, USER_INFO_PATH, "GET", token,
        _identity_query(region, "", now_ms, platform), None, now_ms,
    )
    payload = transport(request, timeout)
    if not isinstance(payload, dict):
        raise CheckinError("Invalid identity response")
    _check_base_resp(payload)
    data = payload.get("data") if isinstance(payload.get("data"), dict) else None
    user_info = None
    for candidate in (
        data.get("userInfo") if data else None,
        data.get("user_info") if data else None,
        payload.get("userInfo"),
        payload.get("user_info"),
    ):
        if isinstance(candidate, dict):
            user_info = candidate
            break
    if user_info is None:
        raise CheckinError("Identity response has no user info")
    for key in ("realUserID", "real_user_id"):
        value = user_info.get(key)
        if isinstance(value, str) and value.strip():
            return value.strip()
    raise CheckinError("Identity response has no account ID")


def summarize_panel(panel) -> dict:
    if not isinstance(panel, dict) or not isinstance(panel.get("days"), list):
        raise CheckinError("Invalid sign-in panel")
    days = []
    for day in panel["days"]:
        if not isinstance(day, dict):
            raise CheckinError("Invalid sign-in panel day")
        days.append({
            "day_no": day.get("day_no"),
            "points": day.get("points"),
            "status": day.get("status"),
            "is_today": day.get("is_today"),
        })
    days.sort(key=lambda item: item["day_no"] if isinstance(item["day_no"], int) else 0)
    return {"scene": panel.get("scene"), "days": days}


def run(credential, claim=False, transport=None, origin=None, region=None, user_id=None,
        app_version=DEFAULT_APP_VERSION, platform=None, now_ms=None, timeout=20.0):
    """Query status and optionally claim the daily sign-in.

    `transport(request, timeout) -> dict` is injectable for offline tests.
    """
    if not isinstance(credential, dict):
        raise CheckinError("Invalid credential object")
    transport = transport or request_json
    token = credential.get("access_token", credential.get("accessToken"))
    if not isinstance(token, str) or not token.strip():
        raise CheckinError("Missing access_token")
    token = token.strip()

    if region is None:
        region = credential.get("region")
    if region not in ORIGINS:
        region = infer_region(credential, origin)
    origin = (origin or ORIGINS[region]).rstrip("/")
    if origin not in ALLOWED_ORIGINS:
        raise CheckinError("Untrusted API origin")
    platform = platform or sys.platform
    now_ms = now_ms if now_ms is not None else int(time.time() * 1000)

    if not user_id:
        user_id = resolve_user_id(origin, token, region, transport, now_ms, app_version, platform, timeout)

    result = {"dryRun": not claim, "region": region, "user_id": user_id}
    status_request = build_request(
        origin, STATUS_PATH, "GET", token,
        _query(region, user_id, now_ms, app_version, platform), None, now_ms,
    )
    panel = summarize_panel(_data(transport(status_request, timeout)))
    result["panel"] = panel
    claimed_today = any(day.get("is_today") and day.get("status") == DAY_CLAIMED for day in panel["days"])
    claimable = any(day.get("status") == DAY_CLAIMABLE for day in panel["days"])
    result["checked_in_today"] = claimed_today

    if not claim:
        return result
    if claimed_today:
        result["claimed"] = False
        result["reason"] = "already-checked-in"
        return result
    if not claimable:
        result["claimed"] = False
        result["reason"] = "unavailable"
        return result

    claim_request = build_request(
        origin, CLAIM_PATH, "POST", token,
        _query(region, user_id, now_ms, app_version, platform), "{}", now_ms,
    )
    claim_data = _data(transport(claim_request, timeout))
    if not isinstance(claim_data, dict):
        raise CheckinError("Invalid claim response")
    claim_result = claim_data.get("claim_result")
    if claim_result == CLAIM_ALREADY_CLAIMED:
        result["claimed"] = False
        result["reason"] = "already-checked-in"
    elif claim_result == CLAIM_CLAIMED:
        result["claimed"] = True
        result["day_no"] = claim_data.get("day_no")
        result["points"] = claim_data.get("points")
        result["expire_at_ms"] = claim_data.get("expire_at_ms")
        if isinstance(claim_data.get("panel"), dict):
            result["panel"] = summarize_panel(claim_data["panel"])
    else:
        raise CheckinError("Unexpected claim result")
    return result


def infer_region(credential: dict, explicit_origin: str | None) -> str:
    if explicit_origin:
        if "minimaxi.com" in explicit_origin or ".minimax.cn" in explicit_origin:
            return "cn"
        return "en"
    base_url = credential.get("base_url") or ""
    if isinstance(base_url, str) and ("minimaxi.com" in base_url or ".minimax.cn" in base_url):
        return "cn"
    return "en"


def load_session(path: Path) -> dict:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, ValueError, UnicodeError):
        raise CheckinError("Cannot read or parse the credential JSON file") from None
    if not isinstance(data, dict):
        raise CheckinError("Invalid credential object")
    # Do not send credentials belonging to other providers to MiniMax.
    provider = data.get("type")
    if provider not in (None, "minimax", "minimax-cn"):
        raise CheckinError("Credential type is not minimax")
    return data


def _auth_dir_files(directory: Path, prefix=None):
    if not directory.is_dir():
        raise CheckinError("--auth-dir is not a directory")
    files = sorted(p for p in directory.glob("*.json")
                   if p.is_file() and (prefix is None or p.name.startswith(prefix)))
    if not files:
        raise CheckinError("No matching .json credential files")
    return files


def _format_result(result: dict) -> str:
    lines = []
    panel = result.get("panel", {})
    days = panel.get("days", [])
    marks = " ".join(
        "✓" if day.get("status") == DAY_CLAIMED else ("·" if day.get("status") != DAY_DISABLED else "x")
        for day in days
    )
    lines.append(f"  region: {result.get('region')}  user_id: {result.get('user_id')}")
    if marks:
        lines.append(f"  cycle: {marks}")
    if result.get("dryRun"):
        lines.append("  query only.")
    elif result.get("claimed"):
        lines.append(f"  checked in: day {result.get('day_no')}, +{result.get('points')} credits")
    elif result.get("reason") == "already-checked-in":
        lines.append("  already checked in today.")
    elif result.get("reason") == "unavailable":
        lines.append("  no claimable day right now.")
    return "\n".join(lines)


def main(argv=None):
    parser = argparse.ArgumentParser(description="MiniMax Code daily sign-in (default: checkin).")
    parser.add_argument("action", nargs="?", default="checkin", choices=["status", "checkin"])
    source = parser.add_mutually_exclusive_group()
    source.add_argument("--auth-file", "--auth", dest="auth_file", type=Path)
    source.add_argument("--auth-dir", type=Path, help="Process matching .json files sequentially")
    source.add_argument("--token", help="Manual access token (prefer a file to avoid shell history)")
    parser.add_argument("--prefix", help="Filename prefix filter for --auth-dir")
    parser.add_argument("--region", choices=sorted(ORIGINS), help="Force en or cn (default: from credential)")
    parser.add_argument("--origin", help="Override the gateway origin (must be allowlisted)")
    parser.add_argument("--user-id", help="Account realUserID (skips the identity lookup)")
    parser.add_argument("--app-version", default=DEFAULT_APP_VERSION, help="Client desktop_version value")
    parser.add_argument("--timeout", type=float, default=20.0)
    parser.add_argument("--claim", action="store_true", help="Legacy alias for checkin")
    parser.add_argument("--json", action="store_true", help="Output JSON")
    args = parser.parse_args(argv)

    if args.prefix is not None and args.auth_dir is None:
        parser.error("--prefix requires --auth-dir")
    if args.user_id is not None and args.token is None and args.auth_file is None and args.auth_dir is None:
        parser.error("--user-id requires --token, --auth-file or --auth-dir")
    if args.claim and args.action == "status":
        parser.error("status cannot be combined with --claim")

    if args.auth_file is None and args.auth_dir is None and args.token is None:
        env_file = os.environ.get("MINIMAX_AUTH_FILE")
        if not env_file:
            parser.error("Use --auth-file, --auth-dir, --token or MINIMAX_AUTH_FILE")
        args.auth_file = Path(env_file)

    try:
        files = _auth_dir_files(args.auth_dir, args.prefix) if args.auth_dir else [args.auth_file]
    except CheckinError as error:
        print(f"Error: {error}", file=sys.stderr)
        return 1

    entries = []
    failed = 0
    for path in files:
        entry: dict = {"source": str(path) if path else "manual token"}
        try:
            credential = load_session(path) if path else {"access_token": args.token}
            result = run(
                credential,
                claim=args.action == "checkin",
                origin=args.origin,
                region=args.region,
                user_id=args.user_id,
                app_version=args.app_version,
                timeout=args.timeout,
            )
            entry["result"] = result
        except CheckinError as error:
            entry["error"] = str(error)
            failed += 1
        except Exception:
            entry["error"] = "Operation failed; no automatic retry performed"
            failed += 1
        entries.append(entry)
        if not args.json:
            print(f"[i] Credential: {entry['source']}")
            if "error" in entry:
                print(f"[x] {entry['error']}")
                continue
            print(_format_result(entry["result"]))

    summary = {"total": len(entries), "succeeded": len(entries) - failed, "failed": failed}
    if args.json:
        print(json.dumps({"accounts": entries, "summary": summary}, ensure_ascii=False, indent=2))
    else:
        print(f"Total: {summary['total']}; succeeded: {summary['succeeded']}; failed: {failed}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
