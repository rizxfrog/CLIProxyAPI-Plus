#!/usr/bin/env python3
"""
WorkBuddy / CodeBuddy 国际版（International Site）模型列表查询脚本（只读，无副作用）

接口与 workbuddy_models_cn.py 完全同构，只有 host / realm / 产品线标识不同
（证据同样来自 reversewk/codebuddy/cli/package 的 dist/codebuddy.js
CloudProductProvider 逆向）:

  - 模型目录: GET {endpoint}/v3/config?repos=<repo,repo,...>
      data 为整份云产品配置，models 只是其中一个字段；repos 参数可省略。
  - 凭证有效性: GET {endpoint}/v2/plugin/account
  - 企业级目录（默认不调用）:
      GET {endpoint}/console/enterprises/{enterpriseId}/config/models
      个人账号 HTTP 403 {"error":"Token does not contain any enterprise id"}。

Host 对照（product.json / officialEndpoints）:
  国际版  https://www.codebuddy.ai            （product.json endpoint）
  预发    https://staging-codebuddy.tencent.com
  国内版  https://copilot.tencent.com         （见 workbuddy_models_cn.py）
  国际版 token 的 realm 是 https://www.codebuddy.ai/auth/realms/copilot，与国内
  https://www.codebuddy.cn/auth/realms/copilot 不同，两站凭据互不通用
  （跨站调用 /v2/plugin/account 一律 401）。

两个必须知道的陷阱（与国内版相同，但目录内容不一样）:

  1) User-Agent 决定服务端返回哪一套目录，缺失或格式不对直接 HTTP 400
     {"code":12403,"msg":"check ua, get coding copilot version error"}。
     实测国际站:
       --product cli -> "CLI/<ver> CodeBuddy/<ver>"   28 条，agents 含
                        cli + general-purpose/compact/contextSummary/… 等内部
                        子任务 agent（国内站只有 1 个 cli agent）
       --product ide -> "WorkBuddy/<ver>"            23 条，只有 cli agent
     国际站 IDE UA 返回的不是 IDE 目录而是更小的 CLI 目录，与国内站行为不同，
     所以国际版请固定用默认 --product cli。
  2) token 失效时 /v3/config 不回 401，而是 HTTP 200 + data.models == null。
     脚本先查 /v2/plugin/account 判定凭证，models 为 null 一律按「凭证无效」报错。

鉴权:
  - Authorization: Bearer <access_token>
  - X-Domain: <endpoint host>（可选，缺省由 endpoint 推导）

登录态来源（自动读取，也可手动传入）:
  命令行:  --auth-file <FILE> / --auth-dir <DIR> [--prefix codebuddy-ai] / --token <JWT>
  环境变量: CODEBUDDY_AUTH_FILE / CODEBUDDY_TOKEN
  auth 文件为本仓库凭据格式: {"type":"codebuddy-ai","access_token":"<JWT>",
  "refresh_token":"...","base_url":"https://www.codebuddy.ai/v2","expired":"..."}

用法:
  uv run docs/scripts/workbuddy_models_intl.py models --auth-dir data/auth_files --prefix codebuddy-ai
  uv run docs/scripts/workbuddy_models_intl.py check  --auth-file data/auth_files/codebuddy-ai-1.json
  uv run docs/scripts/workbuddy_models_intl.py models --token <JWT> --json
  uv run docs/scripts/workbuddy_models_intl.py models --auth-dir data/auth_files --scope all
  uv run docs/scripts/workbuddy_models_intl.py models --auth-dir data/auth_files \
      --out-file internal/registry/models/models.json
      # --out-file: 将 CLI 模型列表合并写入注册表共享目录 models.json 的
      # "codebuddy-ai" 字段（仅 chat 行，剔除图像/视频生成模型）。
      # 按 key 合并：只替换 "codebuddy-ai" 字段，其它 provider 字段原样保留。
      # 注意 --out-file 是整段替换：上游 agents[cli].models 不再列出的 id 会被丢弃。
      # 默认 scope=cli 也是有意的 —— data.models[] 比 CLI 实际对外提供的集合更宽
      # （多出图像/视频生成行，以及只挂在内部子 agent 上的行）。
      # 国际站实测提醒：现有 models.json 的 "codebuddy-ai" 段含 16 个 CLI 目录已不
      # 再列出的 id（claude-*/gemini-*/gpt-5.x-codex 等），整段替换会把它们抹掉；
      # 这些 id 是否仍可用无法从本机验证 —— 国际网关对任何 model 值（含乱写的 id）
      # 都先回 code 11128 "request illegal"，探测法在国际站不成立（国内站才会回
      # code 11102 "service info not found"）。先跑 --json 比对再决定是否写入。
"""
import argparse
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

PROD_ENDPOINT = "https://www.codebuddy.ai"
ALLOWED_ENDPOINT_HOSTS = frozenset(
    {
        "www.codebuddy.ai",
        "codebuddy.ai",
        "staging-codebuddy.tencent.com",
    }
)

CONFIG_PATH = "/v3/config"
ACCOUNT_PATH = "/v2/plugin/account"
ENTERPRISE_MODELS_PATH = "/console/enterprises/{enterprise_id}/config/models"

OK_CODE = 0
UA_ERROR_CODE = 12403

PROVIDER_TYPE = "codebuddy-ai"

# User-Agent selects which catalog the server returns; the version is opaque to
# the gateway (any x.y.z parses). The international gateway answers the same
# "CLI/… CodeBuddy/…" template as the CN one.
PRODUCT_USER_AGENTS = {
    "cli": "CLI/{version} CodeBuddy/{version}",
    "ide": "WorkBuddy/{version}",
}
DEFAULT_PRODUCT_VERSIONS = {"cli": "2.63.2", "ide": "5.3.14"}

# The CLI only advertises the model ids under data.agents[name=="cli"]; the rest
# of data.models belongs to other surfaces (image/video and the extra internal
# sub-agent models the international catalog exposes).
CLI_AGENT_NAME = "cli"

# Tags that mark a non-chat generation model (upstream tags are stable across
# both editions: text-to-image / image-to-image / text-to-video / image-to-video).
GENERATION_TAGS = frozenset(
    {"text-to-image", "image-to-image", "text-to-video", "image-to-video"}
)

# Registry shared-catalog section (internal/registry/models/models.json, key
# "codebuddy-ai"). Field shape mirrors the existing entries in that section:
# object "model", created 1756684800, owned_by "codebuddy", type "codebuddy-ai",
# display_name from the upstream name, description "… via CodeBuddy AI".
REGISTRY_KEY = "codebuddy-ai"
REGISTRY_CREATED = 1756684800
REGISTRY_OWNED_BY = "codebuddy"
REGISTRY_DESCRIPTION_SUFFIX = "via CodeBuddy AI"


class WorkBuddyError(Exception):
    """A sanitized error safe to display without exposing credentials."""


class NoRedirect(urllib.request.HTTPRedirectHandler):
    """Never forward the bearer token through a redirect."""

    def redirect_request(self, req, fp, code, msg, headers, newurl):  # noqa: D102
        return None


def normalize_endpoint(base_url: str) -> str:
    """Return scheme://host for a credential base_url or an explicit endpoint.

    Credentials store base_url with a /v2 suffix (chat gateway); /v3/config and
    /v2/plugin/account are addressed from the origin only.
    """
    raw = str(base_url or "").strip()
    if not raw:
        raise WorkBuddyError("Endpoint is empty")
    if "://" not in raw:
        raw = f"https://{raw}"
    parsed = urllib.parse.urlparse(raw)
    if parsed.scheme != "https":
        raise WorkBuddyError("Endpoint must be https")
    if not parsed.netloc:
        raise WorkBuddyError("Endpoint has no host")
    return f"{parsed.scheme}://{parsed.netloc}"


def validate_endpoint(endpoint: str, allow_custom: bool) -> str:
    host = urllib.parse.urlparse(endpoint).hostname
    if not allow_custom and host not in ALLOWED_ENDPOINT_HOSTS:
        raise WorkBuddyError(
            "Endpoint host is not allowlisted (use --allow-custom-endpoint to override)"
        )
    return endpoint


def build_user_agent(product: str, version: str | None) -> str:
    template = PRODUCT_USER_AGENTS[product]
    return template.format(version=(version or DEFAULT_PRODUCT_VERSIONS[product]).strip())


def request_json(endpoint: str, path: str, token: str, query: dict | None = None,
                 user_agent: str = "", extra_headers: dict | None = None,
                 timeout: float = 15.0) -> dict:
    """GET one gateway endpoint and return the decoded envelope."""
    params = {k: str(v) for k, v in (query or {}).items() if v is not None}
    qs = urllib.parse.urlencode(params)
    url = f"{endpoint}{path}{f'?{qs}' if qs else ''}"
    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/json",
        "X-Domain": urllib.parse.urlparse(endpoint).netloc,
        "User-Agent": user_agent,
    }
    if extra_headers:
        headers.update(extra_headers)
    opener = urllib.request.build_opener(NoRedirect())
    try:
        with opener.open(urllib.request.Request(url, headers=headers), timeout=timeout) as resp:
            status = resp.status
            body = resp.read().decode("utf-8", "replace")
    except urllib.error.HTTPError as error:
        if error.code in (401, 403):
            raise WorkBuddyError(f"Auth rejected by gateway (HTTP {error.code})") from None
        if error.code == 400:
            raise WorkBuddyError(
                "Gateway rejected the request (HTTP 400); check the User-Agent product line"
            ) from None
        raise WorkBuddyError(f"HTTP {error.code} from upstream") from None
    except (urllib.error.URLError, TimeoutError, OSError):
        raise WorkBuddyError("Network request failed; no automatic retry performed") from None
    if status != 200:
        raise WorkBuddyError(f"Unexpected HTTP status {status}")
    try:
        payload = json.loads(body)
    except ValueError:
        raise WorkBuddyError("Upstream returned invalid JSON") from None
    if not isinstance(payload, dict):
        raise WorkBuddyError("Upstream returned an unexpected envelope")
    code = payload.get("code")
    if isinstance(code, str):
        code = int(code) if code.lstrip("-").isdigit() else None
    if code not in (None, OK_CODE):
        message = str(payload.get("msg") or code)
        if code == UA_ERROR_CODE:
            raise WorkBuddyError(f"User-Agent rejected ({code}): {message}")
        if code in (401, 403):
            raise WorkBuddyError(f"Gateway authorization failed ({code}): {message}")
        raise WorkBuddyError(f"Business error {code}: {message}")
    return payload


def check_account(endpoint: str, token: str, user_agent: str, timeout: float = 15.0) -> dict:
    """Probe credential validity; the catalog endpoint cannot tell us."""
    payload = request_json(endpoint, ACCOUNT_PATH, token, user_agent=user_agent,
                           timeout=timeout)
    data = payload.get("data")
    if not isinstance(data, dict):
        raise WorkBuddyError("Account response has no data object")
    return {
        "uid": str(data.get("uid") or ""),
        "uin": str(data.get("uin") or ""),
        "nickname": str(data.get("nickname") or ""),
        "type": str(data.get("type") or ""),
        "enterprise_id": str(data.get("enterpriseId") or ""),
        "plugin_enabled": bool(data.get("pluginEnabled")),
        "phone": str(data.get("phoneNumber") or ""),
    }


def fetch_config(endpoint: str, token: str, user_agent: str, repos: list[str],
                 timeout: float = 15.0) -> dict:
    """Fetch the cloud product configuration and return data.models rows."""
    payload = request_json(endpoint, CONFIG_PATH, token, query={"repos": ",".join(repos)},
                           user_agent=user_agent, timeout=timeout)
    data = payload.get("data")
    if not isinstance(data, dict):
        raise WorkBuddyError("Configuration response has no data object")
    rows = data.get("models")
    if rows is None:
        # Anonymous/revoked credentials answer HTTP 200 with models == null.
        raise WorkBuddyError(
            "Catalog is null: the credential is missing, expired or revoked "
            "(run the 'check' action to confirm; re-login required)"
        )
    if not isinstance(rows, list):
        raise WorkBuddyError("Configuration response has no model list")
    return {"data": data, "models": [row for row in rows if isinstance(row, dict)
                                     and str(row.get("id") or "").strip()]}


def fetch_enterprise_models(endpoint: str, token: str, user_agent: str,
                            enterprise_id: str, timeout: float = 15.0) -> list[dict]:
    """Enterprise-scoped catalog (ModelsProductProvider); personal accounts 403."""
    path = ENTERPRISE_MODELS_PATH.format(enterprise_id=urllib.parse.quote(enterprise_id))
    payload = request_json(endpoint, path, token, user_agent=user_agent, timeout=timeout)
    data = payload.get("data")
    if isinstance(data, dict):
        data = data.get("data")
    if not isinstance(data, list):
        raise WorkBuddyError("Enterprise models response has no list")
    return [row for row in data if isinstance(row, dict) and str(row.get("id") or "").strip()]


def agent_ids(data: dict, agent_name: str) -> list[str]:
    """Model ids advertised by one agent; agents[].models holds plain id strings."""
    for agent in data.get("agents") or []:
        if not isinstance(agent, dict) or str(agent.get("name")) != agent_name:
            continue
        ids = []
        for item in agent.get("models") or []:
            if isinstance(item, str) and item.strip():
                ids.append(item.strip())
            elif isinstance(item, dict) and str(item.get("id") or "").strip():
                ids.append(str(item["id"]).strip())
        return ids
    return []


def agent_names(data: dict) -> list[str]:
    return [str(agent.get("name")) for agent in (data.get("agents") or [])
            if isinstance(agent, dict)]


def select_rows(result: dict, scope: str) -> list[dict]:
    """Resolve --scope into upstream model rows, preserving catalog order."""
    rows = result["models"]
    if scope == "all":
        return rows
    by_id = {str(row.get("id")): row for row in rows}
    if scope == "cli":
        wanted = agent_ids(result["data"], CLI_AGENT_NAME)
    else:
        wanted = agent_ids(result["data"], scope.removeprefix("agent="))
    if not wanted:
        raise WorkBuddyError(f"Scope '{scope}' is not advertised by this catalog "
                             f"(agents: {', '.join(agent_names(result['data'])) or 'none'})")
    selected = [by_id[mid] for mid in wanted if mid in by_id]
    missing = [mid for mid in wanted if mid not in by_id]
    if missing:
        raise WorkBuddyError(f"Scope '{scope}' references unknown models: {', '.join(missing)}")
    return selected


def is_generation_model(row: dict) -> bool:
    tags = {str(tag).strip().lower() for tag in (row.get("tags") or [])}
    return bool(tags & GENERATION_TAGS)


def thinking_from(row: dict) -> dict | None:
    """Map the upstream reasoning block onto the registry ThinkingSupport shape."""
    reasoning = row.get("reasoning") if isinstance(row.get("reasoning"), dict) else {}
    levels = [str(level).strip() for level in (reasoning.get("supportedEfforts") or [])
              if str(level).strip()]
    if not levels:
        single = str(reasoning.get("effort") or "").strip()
        levels = [single] if single else []
    if not levels:
        return None
    return {
        "zero_allowed": bool(reasoning.get("canDisableThinking")),
        "levels": levels,
    }


def to_registry_catalog(rows: list[dict]) -> dict:
    """Convert raw /v3/config model rows into the embedded fallback-catalog shape."""
    entries = []
    seen = set()
    for row in rows:
        model_id = str(row.get("id") or "").strip()
        if not model_id or model_id in seen or is_generation_model(row):
            continue
        seen.add(model_id)
        name = str(row.get("name") or "").strip() or model_id
        entry: dict = {
            "id": model_id,
            "object": "model",
            "created": REGISTRY_CREATED,
            "owned_by": REGISTRY_OWNED_BY,
            "type": REGISTRY_KEY,
            "display_name": name,
            "description": f"{name} {REGISTRY_DESCRIPTION_SUFFIX}",
        }
        context = row.get("maxInputTokens")
        if isinstance(context, (int, float)) and not isinstance(context, bool) and context > 0:
            entry["context_length"] = int(context)
        output = row.get("maxOutputTokens")
        if isinstance(output, (int, float)) and not isinstance(output, bool) and output > 0:
            entry["max_completion_tokens"] = int(output)
        thinking = thinking_from(row)
        if thinking:
            entry["thinking"] = thinking
        modalities = ["text", "image"] if row.get("supportsImages") else ["text"]
        entry["supportedInputModalities"] = modalities
        entry["supportedOutputModalities"] = ["text"]
        entries.append(entry)
    return {REGISTRY_KEY: entries}


def write_registry_catalog(catalog: dict, out_path: Path) -> None:
    """Merge catalog[REGISTRY_KEY] into the shared registry catalog by key.

    Other provider sections in an existing file are preserved; only the
    "codebuddy-ai" section is replaced. A missing file is created. A file that
    exists but is not a readable JSON object raises, so the catalog is
    never clobbered.
    """
    document: dict = {}
    if out_path.exists():
        try:
            existing = json.loads(out_path.read_text(encoding="utf-8"))
        except (OSError, ValueError, UnicodeError):
            raise WorkBuddyError("--out-file exists but is not readable JSON; refusing to overwrite") from None
        if not isinstance(existing, dict):
            raise WorkBuddyError("--out-file is not a JSON object; refusing to overwrite")
        document = existing
    document[REGISTRY_KEY] = catalog[REGISTRY_KEY]
    try:
        parent = out_path.parent
        if str(parent) not in ("", "."):
            parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(
            json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    except OSError:
        raise WorkBuddyError("Cannot write --out-file") from None


def load_credential(path: Path) -> dict:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, ValueError, UnicodeError):
        raise WorkBuddyError("Cannot read or parse the credential JSON file") from None
    if not isinstance(data, dict):
        raise WorkBuddyError("Invalid credential object")
    provider = data.get("type")
    if provider not in (None, PROVIDER_TYPE):
        raise WorkBuddyError(f"Credential type is not {PROVIDER_TYPE}")
    return data


def _auth_dir_files(directory: Path, prefix=None):
    if not directory.is_dir():
        raise WorkBuddyError("--auth-dir is not a directory")
    files = sorted(p for p in directory.glob("*.json")
                   if p.is_file() and (prefix is None or p.name.startswith(prefix)))
    if not files:
        raise WorkBuddyError("No matching .json credential files")
    return files


def resolve_token(credential: dict) -> str:
    token = str(credential.get("access_token") or credential.get("token") or "").strip()
    if not token:
        raise WorkBuddyError("Credential has no access_token")
    return token


def resolve_endpoint(credential: dict, explicit: str | None) -> str:
    if explicit:
        return explicit
    base = str(credential.get("base_url") or "").strip()
    if base:
        return base
    return PROD_ENDPOINT


def _fmt_model(row: dict) -> str:
    parts = [str(row.get("id"))]
    name = str(row.get("name") or "").strip()
    if name and name != row.get("id"):
        parts.append(f"({name})")
    meta = []
    if isinstance(row.get("maxInputTokens"), (int, float)):
        meta.append(f"context={row['maxInputTokens']}")
    if isinstance(row.get("maxOutputTokens"), (int, float)):
        meta.append(f"output={row['maxOutputTokens']}")
    reasoning = row.get("reasoning") if isinstance(row.get("reasoning"), dict) else {}
    efforts = reasoning.get("supportedEfforts") or ([reasoning["effort"]] if reasoning.get("effort") else [])
    if efforts:
        meta.append("reasoning[" + ",".join(str(e) for e in efforts) + "]")
    if row.get("supportsToolCall"):
        meta.append("tools")
    if row.get("supportsImages"):
        meta.append("vision")
    tags = [str(tag) for tag in (row.get("tags") or []) if str(tag).strip()]
    if tags:
        meta.append("tags=" + ",".join(tags))
    credits = str(row.get("credits") or "").strip()
    if credits:
        meta.append(credits)
    if meta:
        parts.append("[" + ", ".join(meta) + "]")
    return "  ".join(parts)


def _format_models(models: list[dict], scope: str, product: str) -> str:
    lines = [f"  product={product} scope={scope}: {len(models)} model(s):"]
    lines.extend(f"  - {_fmt_model(item)}" for item in models)
    return "\n".join(lines)


def _format_account(account: dict) -> str:
    return "\n".join([
        f"  uid:            {account.get('uid') or '--'}",
        f"  uin:            {account.get('uin') or '--'}",
        f"  nickname:       {account.get('nickname') or '--'}",
        f"  account type:   {account.get('type') or '--'}"
        + ("" if account.get("type") else " (unknown)"),
        f"  enterpriseId:   {account.get('enterprise_id') or '-- (personal: /console/.../config/models is unavailable)'}",
        f"  pluginEnabled:  {account.get('plugin_enabled')}",
        f"  phone:          {account.get('phone') or '--'}",
    ])


def run_one(action: str, credential: dict, args, user_agent: str,
            timeout: float) -> dict:
    token = resolve_token(credential)
    endpoint = validate_endpoint(normalize_endpoint(resolve_endpoint(credential, args.endpoint)),
                                 args.allow_custom_endpoint)
    if action == "check":
        return {"endpoint": endpoint, "user_agent": user_agent,
                "account": check_account(endpoint, token, user_agent, timeout=timeout)}
    if action == "enterprise":
        account = check_account(endpoint, token, user_agent, timeout=timeout)
        enterprise_id = args.enterprise_id or account.get("enterprise_id")
        if not enterprise_id:
            raise WorkBuddyError(
                "No enterpriseId for this credential (personal account); "
                "the enterprise catalog is not available")
        rows = fetch_enterprise_models(endpoint, token, user_agent, enterprise_id,
                                       timeout=timeout)
        return {"endpoint": endpoint, "user_agent": user_agent,
                "enterprise_id": enterprise_id, "count": len(rows), "models": rows}
    repos = [str(repo).strip() for repo in (args.repos or []) if str(repo).strip()]
    result = fetch_config(endpoint, token, user_agent, repos, timeout=timeout)
    models = select_rows(result, args.scope)
    payload = {"endpoint": endpoint, "user_agent": user_agent, "scope": args.scope,
               "product": args.product, "count": len(models), "models": models}
    if args.raw:
        # Whole unfiltered catalog: --scope only narrows what is advertised.
        payload["catalog_count"] = len(result["models"])
        payload["catalog_models"] = result["models"]
        payload["agents"] = {name: agent_ids(result["data"], name)
                             for name in agent_names(result["data"])}
    return payload


def main(argv=None):
    parser = argparse.ArgumentParser(
        description="WorkBuddy/CodeBuddy international model catalog from /v3/config (read-only).")
    parser.add_argument("action", choices=["models", "check", "enterprise"],
                        help="models: catalog; check: credential/account probe; "
                             "enterprise: enterprise-scoped catalog (enterprise accounts only)")
    source = parser.add_mutually_exclusive_group()
    source.add_argument("--auth-file", "--auth", dest="auth_file", type=Path)
    source.add_argument("--auth-dir", type=Path, help="Process matching .json files sequentially")
    source.add_argument("--token", help="Manual bearer token (prefer a file to avoid shell history)")
    parser.add_argument("--prefix", help="Filename prefix filter for --auth-dir")
    parser.add_argument("--endpoint", help=f"Override endpoint (default: credential base_url or {PROD_ENDPOINT})")
    parser.add_argument("--allow-custom-endpoint", action="store_true", help="Skip the host allowlist check")
    parser.add_argument("--product", choices=sorted(PRODUCT_USER_AGENTS), default="cli",
                        help="User-Agent product line; selects which catalog the gateway returns")
    parser.add_argument("--version", help="Version used in the User-Agent (opaque to the gateway)")
    parser.add_argument("--scope", default="cli",
                        help="cli (default, what the CLI advertises) | all | agent=<name>")
    parser.add_argument("--repos", nargs="*", default=[],
                        help="Optional repos query values (the catalog is identical without them)")
    parser.add_argument("--enterprise-id", help="enterpriseId for the enterprise action")
    parser.add_argument("--out-file", type=Path,
                        help="Merge the model list into the registry shared catalog "
                             "(internal/registry/models/models.json, 'codebuddy-ai' section); "
                             "other provider sections are preserved")
    parser.add_argument("--raw", action="store_true", help="Include the upstream rows verbatim in --json output")
    parser.add_argument("--timeout", type=float, default=15.0)
    parser.add_argument("--json", action="store_true", help="Output JSON")
    args = parser.parse_args(argv)

    if args.prefix is not None and args.auth_dir is None:
        parser.error("--prefix requires --auth-dir")
    if args.out_file is not None and args.action != "models":
        parser.error("--out-file requires the models action")
    if args.out_file is not None and args.scope != "cli":
        parser.error("--out-file requires the default cli scope (the registry catalog "
                     "mirrors what the CLI advertises)")
    if args.enterprise_id is not None and args.action != "enterprise":
        parser.error("--enterprise-id requires the enterprise action")

    user_agent = build_user_agent(args.product, args.version)

    if args.auth_file is None and args.auth_dir is None and args.token is None:
        env_file = os.environ.get("CODEBUDDY_AUTH_FILE")
        env_token = os.environ.get("CODEBUDDY_TOKEN")
        if env_file:
            args.auth_file = Path(env_file)
        elif env_token:
            args.token = env_token
        else:
            parser.error("Use --auth-file, --auth-dir, --token or CODEBUDDY_AUTH_FILE/CODEBUDDY_TOKEN")

    try:
        files = _auth_dir_files(args.auth_dir, args.prefix) if args.auth_dir else [args.auth_file]
    except WorkBuddyError as error:
        print(f"Error: {error}", file=sys.stderr)
        return 1

    entries = []
    failed = 0
    for path in files:
        entry: dict = {"source": str(path) if path else "manual token"}
        try:
            credential = load_session_or_manual(path, args.token)
            result = run_one(args.action, credential, args, user_agent, args.timeout)
            if args.out_file:
                catalog = to_registry_catalog(result.get("models", []))
                if not catalog[REGISTRY_KEY]:
                    raise WorkBuddyError("No chat models returned; --out-file not written")
                write_registry_catalog(catalog, args.out_file)
                result["out_file"] = str(args.out_file)
                result["registry_count"] = len(catalog[REGISTRY_KEY])
            entry["result"] = result
        except WorkBuddyError as error:
            entry["error"] = str(error)
            failed += 1
        except Exception:
            entry["error"] = "Operation failed; no automatic retry performed"
            failed += 1
        entries.append(entry)
        if not args.json:
            print(f"[i] Credential: {entry['source']}  (UA: {user_agent})")
            if "error" in entry:
                print(f"[x] {entry['error']}")
                continue
            if args.action == "check":
                print(_format_account(entry["result"]["account"]))
            else:
                print(_format_models(entry["result"]["models"], entry["result"]["scope"],
                                     entry["result"]["product"]))
                if entry["result"].get("out_file"):
                    print(f"  registry catalog written: {entry['result']['out_file']} "
                          f"({entry['result']['registry_count']} chat entries)")

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
