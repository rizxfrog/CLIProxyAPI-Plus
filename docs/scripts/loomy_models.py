#!/usr/bin/env python3
"""
Loomy（讯飞 Loomy）模型列表 / 积分额度查询脚本（只读，无副作用）

接口（从 Loomy-0.9.39-win 安装包 resources/app.asar 逆向提取，
逐层证据见 docs/providers/loomy-integration.md §6.1 / §7.1）:

  - 模型列表: GET {base}/models?pageNo=1&pageSize=100
      客户端实现 electron/model-service.js fetchModelsFromProvider 不带分页参数；
      服务端 envelope 支持分页（与 points/records 同一套 code/desc/data 包装）。
  - 积分额度: GET {base}/points/records?pageNo=1&pageSize=1&recordType=all
      data.balance = 永久/长期积分, data.dailyBalance = 每日赠送积分,
      另可能返回 dailyLimitPoints / dailyConsumedPoints / dailyRemainingPoints
      （客户端按可空解析，缺失即视为未知，不硬编码 5000）。

鉴权:
  - 仅小写请求头 `token: <session>`（32 位 hex 登录会话）。
    注意：/models 与 /points/* 带 `Authorization: Bearer` 反而会回 100002。
  - 业务响应恒为 HTTP 200 信封 {code, desc, data}；code == "000000" 为成功，
    "100002" 表示未登录/登录失效（需重新登录，重试无意义）。

Base URL:
  - 生产: https://loomyad.xunfei.cn/api/v1
    （安装包 resources/.env.prod 解密得 LOOMY_POINTS_BASE_URL=https://loomyad.xunfei.cn；
     客户端运行时由 pointsConfig.baseUrl 强制覆写 provider imodel 的 baseURL 并追加 /api/v1）
  - 测试: https://ossptest.voicecloud.cn/loomy/integration/api/v1
    （安装包内 opencode-home-template / points-config.js 的默认值）

登录态来源（自动读取，也可手动传入）:
  命令行:  --auth-file <FILE> / --auth-dir <DIR> [--prefix loomy] / --token <SESSION>
  环境变量: LOOMY_AUTH_FILE / LOOMY_TOKEN
  auth 文件为本仓库凭据格式: {"type":"loomy","access_token":"<session>",
  "base_url":"...", "userid":"...", "phone":"...", "expired":"..."}

用法:
  uv run docs/scripts/loomy_models.py models --auth-dir auths
  uv run docs/scripts/loomy_models.py quota  --auth-file auths/loomy-xxx.json
  uv run docs/scripts/loomy_models.py models --token <session> --json
  uv run docs/scripts/loomy_models.py models --auth-file auths/loomy-x.json \
      --out-file internal/registry/models/models.json
      # --out-file: 将模型列表合并写入注册表共享目录 models.json 的 "loomy" 字段
      # （仅 type=="chat" 行，口径同 internal/auth/loomy modelFromRow）。
      # 按 key 合并：只替换 "loomy" 字段，其它 provider 字段原样保留。

  uv run docs/scripts/loomy_models.py models --auth-dir auths --out-file internal/registry/models/models.json
"""
import argparse
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

PROD_BASE = "https://loomyad.xunfei.cn/api/v1"
TEST_BASE = "https://ossptest.voicecloud.cn/loomy/integration/api/v1"
ALLOWED_BASE_HOSTS = frozenset(
    {"loomyad.xunfei.cn", "ossptest.voicecloud.cn"}
)

OK_CODE = "000000"
AUTH_ERROR_CODE = "100002"

MODELS_PATH = "/models"
POINTS_PATH = "/points/records"

PROVIDER_TYPE = "loomy"


class LoomyError(Exception):
    """A sanitized error safe to display without exposing credentials."""


class NoRedirect(urllib.request.HTTPRedirectHandler):
    """Never forward the session token through a redirect."""

    def redirect_request(self, req, fp, code, msg, headers, newurl):  # noqa: D102
        return None


def normalize_base(base_url: str) -> str:
    """Trim trailing slashes and enforce the /api/v1 suffix like the client does."""
    base = str(base_url or "").strip().rstrip("/")
    if not base:
        raise LoomyError("Base URL is empty")
    if not base.endswith("/api/v1"):
        base = f"{base}/api/v1"
    return base


def validate_base(base: str, allow_custom: bool) -> str:
    parsed = urllib.parse.urlparse(base)
    if parsed.scheme != "https":
        raise LoomyError("Base URL must be https")
    if not allow_custom and parsed.hostname not in ALLOWED_BASE_HOSTS:
        raise LoomyError("Base URL host is not allowlisted (use --allow-custom-base to override)")
    return base


def build_url(base: str, path: str, query: dict | None = None) -> str:
    params = {k: str(v) for k, v in (query or {}).items() if v is not None and str(v) != ""}
    qs = urllib.parse.urlencode(params)
    return f"{base}{path}{f'?{qs}' if qs else ''}"


def request_json(base: str, path: str, token: str, query: dict | None = None,
                 timeout: float = 15.0) -> dict:
    """GET one business endpoint and return the decoded envelope."""
    url = build_url(base, path, query)
    # The client sends the lowercase token header only; Bearer breaks /models.
    headers = {
        "token": token,
        "Accept": "application/json",
        "Content-Type": "application/json",
    }
    opener = urllib.request.build_opener(NoRedirect())
    try:
        with opener.open(urllib.request.Request(url, headers=headers), timeout=timeout) as resp:
            body = resp.read().decode("utf-8", "replace")
    except urllib.error.HTTPError as error:
        if error.code in (401, 403):
            raise LoomyError(f"Auth rejected by transport (HTTP {error.code})") from None
        raise LoomyError(f"HTTP {error.code} from upstream") from None
    except (urllib.error.URLError, TimeoutError, OSError):
        raise LoomyError("Network request failed; no automatic retry performed") from None
    try:
        payload = json.loads(body)
    except ValueError:
        raise LoomyError("Upstream returned invalid JSON") from None
    if not isinstance(payload, dict):
        raise LoomyError("Upstream returned an unexpected envelope")
    code = str(payload.get("code") or "")
    if code and code != OK_CODE:
        desc = str(payload.get("desc") or code)
        if code == AUTH_ERROR_CODE:
            raise LoomyError(f"Session invalid or expired (100002): {desc}; re-login required")
        raise LoomyError(f"Business error {code}: {desc}")
    return payload


def fetch_models(base: str, token: str, timeout: float = 15.0) -> list[dict]:
    payload = request_json(base, MODELS_PATH, token, timeout=timeout)
    data = payload.get("data")
    if isinstance(data, dict):
        data = data.get("records") or data.get("list") or data.get("data")
    if not isinstance(data, list):
        raise LoomyError("Models response has no list")
    return [item for item in data if isinstance(item, dict) and str(item.get("id") or "").strip()]


def fetch_quota(base: str, token: str, timeout: float = 15.0) -> dict:
    payload = request_json(
        base, POINTS_PATH, token,
        query={"pageNo": 1, "pageSize": 1, "recordType": "all"},
        timeout=timeout,
    )
    data = payload.get("data")
    if not isinstance(data, dict):
        raise LoomyError("Points response has no data object")

    def num(value):
        try:
            out = float(value)
        except (TypeError, ValueError):
            return None
        return None if out != out else value  # drop NaN

    return {
        "balance": num(data.get("balance")),
        "daily_balance": num(data.get("dailyBalance")),
        "daily_limit_points": num(data.get("dailyLimitPoints")),
        "daily_consumed_points": num(data.get("dailyConsumedPoints")),
        "daily_remaining_points": num(data.get("dailyRemainingPoints")),
        "total": num(data.get("total")),
    }


def load_credential(path: Path) -> dict:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, ValueError, UnicodeError):
        raise LoomyError("Cannot read or parse the credential JSON file") from None
    if not isinstance(data, dict):
        raise LoomyError("Invalid credential object")
    provider = data.get("type")
    if provider not in (None, PROVIDER_TYPE):
        raise LoomyError("Credential type is not loomy")
    return data


def _auth_dir_files(directory: Path, prefix=None):
    if not directory.is_dir():
        raise LoomyError("--auth-dir is not a directory")
    files = sorted(p for p in directory.glob("*.json")
                   if p.is_file() and (prefix is None or p.name.startswith(prefix)))
    if not files:
        raise LoomyError("No matching .json credential files")
    return files


def resolve_token(credential: dict) -> str:
    token = str(credential.get("access_token") or credential.get("token") or "").strip()
    if not token:
        raise LoomyError("Credential has no access_token (session)")
    return token


def resolve_base(credential: dict, explicit: str | None) -> str:
    if explicit:
        return explicit
    base = str(credential.get("base_url") or "").strip()
    if base:
        return base
    return PROD_BASE


def _fmt_model(item: dict) -> str:
    caps = item.get("capabilities") if isinstance(item.get("capabilities"), dict) else {}
    flags = []
    if item.get("reasoning") or caps.get("reasoning"):
        flags.append("reasoning")
    if item.get("tool_call") or caps.get("function_calling"):
        flags.append("tools")
    modalities = caps.get("input_modalities") or []
    if isinstance(modalities, list) and any(m in ("image", "video", "pdf") for m in modalities):
        flags.append("vision")
    protocol = str(item.get("protocol") or "").strip()
    limit = item.get("limit") if isinstance(item.get("limit"), dict) else {}
    context = item.get("context_length") or limit.get("context")
    defaults = []
    if item.get("is_default_model") is True:
        defaults.append("default")
    if item.get("is_default_small_model") is True:
        defaults.append("small-default")
    parts = [str(item.get("id"))]
    name = str(item.get("name") or "").strip()
    if name and name != item.get("id"):
        parts.append(f"({name})")
    meta = []
    if protocol:
        meta.append(protocol)
    if context:
        meta.append(f"context={context}")
    meta.extend(flags)
    meta.extend(defaults)
    if meta:
        parts.append("[" + ", ".join(str(m) for m in meta) + "]")
    return "  ".join(parts)


def _format_models(models: list[dict]) -> str:
    lines = [f"  {len(models)} model(s):"]
    lines.extend(f"  - {_fmt_model(item)}" for item in models)
    return "\n".join(lines)


def _format_quota(quota: dict) -> str:
    def show(value):
        return "--" if value is None else value

    lines = [
        f"  permanent (balance):      {show(quota.get('balance'))}",
        f"  daily gift (dailyBalance): {show(quota.get('daily_balance'))}",
    ]
    optional = [
        ("dailyLimitPoints", quota.get("daily_limit_points")),
        ("dailyConsumedPoints", quota.get("daily_consumed_points")),
        ("dailyRemainingPoints", quota.get("daily_remaining_points")),
    ]
    present = [f"{name}={value}" for name, value in optional if value is not None]
    lines.append("  daily window: " + (", ".join(present) if present else "not returned (unknown, not 0)"))
    return "\n".join(lines)


# Registry shared-catalog section (internal/registry/models/models.json, key
# "loomy"). The mapping mirrors internal/auth/loomy modelFromRow (only
# type=="chat" rows, reasoning_efforts -> thinking.levels, "none" level ->
# zero_allowed) plus sdk/cliproxy loomyCatalogueModelInfo defaults
# (object "model", created 1759276800, owned_by "iflytek", type "loomy").
REGISTRY_KEY = "loomy"
REGISTRY_CREATED = 1759276800


def to_registry_catalog(rows: list[dict]) -> dict:
    """Convert raw /models rows into the embedded fallback-catalog shape."""
    entries = []
    seen = set()
    for row in rows:
        if str(row.get("type") or "").strip() != "chat":
            continue
        model_id = str(row.get("id") or "").strip()
        if not model_id or model_id in seen:
            continue
        seen.add(model_id)
        caps = row.get("capabilities") if isinstance(row.get("capabilities"), dict) else {}
        name = str(row.get("name") or "").strip() or model_id
        entry: dict = {
            "id": model_id,
            "object": "model",
            "created": REGISTRY_CREATED,
            "owned_by": "iflytek",
            "type": "loomy",
            "display_name": name,
        }
        description = str(row.get("description") or "").strip()
        if description:
            entry["description"] = description
        context = row.get("context_length")
        if isinstance(context, (int, float)) and not isinstance(context, bool) and context > 0:
            entry["context_length"] = int(context)
        levels = [str(level).strip() for level in (row.get("reasoning_efforts") or [])
                  if str(level).strip()]
        if levels:
            entry["thinking"] = {
                "zero_allowed": any(level.lower() == "none" for level in levels),
                "levels": levels,
            }
        input_mods = [str(mod).strip().lower() for mod in (caps.get("input_modalities") or [])
                      if str(mod).strip()]
        entry["supportedInputModalities"] = input_mods or ["text"]
        output_mods = [str(mod).strip().lower() for mod in (caps.get("output_modalities") or [])
                       if str(mod).strip()]
        entry["supportedOutputModalities"] = output_mods or ["text"]
        entries.append(entry)
    return {"loomy": entries}


def write_registry_catalog(catalog: dict, out_path: Path) -> None:
    """Merge catalog["loomy"] into the shared registry catalog by key.

    Other provider sections in an existing file are preserved; only the
    "loomy" section is replaced. A missing file is created. A file that
    exists but is not a readable JSON object raises, so the catalog is
    never clobbered.
    """
    document: dict = {}
    if out_path.exists():
        try:
            existing = json.loads(out_path.read_text(encoding="utf-8"))
        except (OSError, ValueError, UnicodeError):
            raise LoomyError("--out-file exists but is not readable JSON; refusing to overwrite") from None
        if not isinstance(existing, dict):
            raise LoomyError("--out-file is not a JSON object; refusing to overwrite")
        document = existing
    document[REGISTRY_KEY] = catalog[REGISTRY_KEY]
    try:
        parent = out_path.parent
        if str(parent) not in ("", "."):
            parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(
            json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    except OSError:
        raise LoomyError("Cannot write --out-file") from None


def run_one(action: str, credential: dict, base_override: str | None,
            allow_custom: bool, timeout: float) -> dict:
    token = resolve_token(credential)
    base = normalize_base(resolve_base(credential, base_override))
    validate_base(base, allow_custom)
    if action == "models":
        models = fetch_models(base, token, timeout=timeout)
        return {"base_url": base, "count": len(models), "models": models}
    quota = fetch_quota(base, token, timeout=timeout)
    return {"base_url": base, "quota": quota}


def main(argv=None):
    parser = argparse.ArgumentParser(description="Loomy model list / points quota (read-only).")
    parser.add_argument("action", choices=["models", "quota"])
    source = parser.add_mutually_exclusive_group()
    source.add_argument("--auth-file", "--auth", dest="auth_file", type=Path)
    source.add_argument("--auth-dir", type=Path, help="Process matching .json files sequentially")
    source.add_argument("--token", help="Manual session token (prefer a file to avoid shell history)")
    parser.add_argument("--prefix", help="Filename prefix filter for --auth-dir")
    parser.add_argument("--base-url", help=f"Override base URL (default: credential or prod {PROD_BASE})")
    parser.add_argument("--test-env", action="store_true", help=f"Use the test gateway {TEST_BASE}")
    parser.add_argument("--allow-custom-base", action="store_true", help="Skip the host allowlist check")
    parser.add_argument("--out-file", type=Path,
                        help="Merge the model list into the registry shared catalog "
                             "(internal/registry/models/models.json, 'loomy' section); "
                             "other provider sections are preserved")
    parser.add_argument("--timeout", type=float, default=15.0)
    parser.add_argument("--json", action="store_true", help="Output JSON")
    args = parser.parse_args(argv)

    if args.prefix is not None and args.auth_dir is None:
        parser.error("--prefix requires --auth-dir")
    if args.base_url and args.test_env:
        parser.error("--base-url and --test-env are mutually exclusive")
    if args.out_file is not None and args.action != "models":
        parser.error("--out-file requires the models action")

    base_override = args.base_url or (TEST_BASE if args.test_env else None)

    if args.auth_file is None and args.auth_dir is None and args.token is None:
        env_file = os.environ.get("LOOMY_AUTH_FILE")
        env_token = os.environ.get("LOOMY_TOKEN")
        if env_file:
            args.auth_file = Path(env_file)
        elif env_token:
            args.token = env_token
        else:
            parser.error("Use --auth-file, --auth-dir, --token or LOOMY_AUTH_FILE/LOOMY_TOKEN")

    try:
        files = _auth_dir_files(args.auth_dir, args.prefix) if args.auth_dir else [args.auth_file]
    except LoomyError as error:
        print(f"Error: {error}", file=sys.stderr)
        return 1

    entries = []
    failed = 0
    for path in files:
        entry: dict = {"source": str(path) if path else "manual token"}
        try:
            credential = load_session_or_manual(path, args.token)
            result = run_one(args.action, credential, base_override,
                             args.allow_custom_base, args.timeout)
            if args.out_file:
                catalog = to_registry_catalog(result.get("models", []))
                if not catalog["loomy"]:
                    raise LoomyError("No chat-type models returned; --out-file not written")
                write_registry_catalog(catalog, args.out_file)
                result["out_file"] = str(args.out_file)
                result["registry_count"] = len(catalog["loomy"])
            entry["result"] = result
        except LoomyError as error:
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
            if args.action == "models":
                print(_format_models(entry["result"]["models"]))
                if entry["result"].get("out_file"):
                    print(f"  registry catalog written: {entry['result']['out_file']} "
                          f"({entry['result']['registry_count']} chat entries)")
            else:
                print(_format_quota(entry["result"]["quota"]))

    summary = {"total": len(entries), "succeeded": len(entries) - failed, "failed": failed}
    if args.json:
        print(json.dumps({"accounts": entries, "summary": summary}, ensure_ascii=False, indent=2))
    else:
        print(f"Total: {summary['total']}; succeeded: {summary['succeeded']}; failed: {failed}")
    return 1 if failed else 0


def load_session_or_manual(path: Path | None, manual_token: str | None) -> dict:
    if path is not None:
        return load_credential(path)
    return {"access_token": (manual_token or "").strip()}


if __name__ == "__main__":
    sys.exit(main())
