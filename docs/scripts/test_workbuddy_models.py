#!/usr/bin/env python3
"""Offline regression tests for workbuddy_models_cn.py / workbuddy_models_intl.py.

Never call the live service. The two scripts share identical logic and differ
only in their edition constants, so the shared behaviour is exercised on the CN
module and the edition-specific constants are asserted on both.
"""
import io
import json
import sys
import tempfile
import unittest
import urllib.error
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent))

import workbuddy_models_cn as cn
import workbuddy_models_intl as intl

TOKEN = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1In0.sig"

CN_ENDPOINT = "https://copilot.tencent.com"
INTL_ENDPOINT = "https://www.codebuddy.ai"

# Upstream-shaped rows: catalog order deliberately differs from the agent order,
# one row carries a generation tag, one has the effort form instead of
# supportedEfforts, one has no reasoning at all.
MODEL_ROWS = [
    {"id": "default", "name": "Default", "maxInputTokens": 200000,
     "maxOutputTokens": 24000, "supportsToolCall": True, "supportsImages": False},
    {"id": "hunyuan-image-alpha", "name": "Hunyuan Image",
     "tags": ["text-to-image"], "supportsToolCall": False},
    {"id": "glm-5.3", "name": "GLM-5.3", "maxInputTokens": 1000000,
     "maxOutputTokens": 64000, "supportsToolCall": True, "supportsImages": True,
     "reasoning": {"canDisableThinking": False, "defaultEffort": "max",
                   "supportedEfforts": ["low", "high", "max"]},
     "credits": "x0.69 credits"},
    {"id": "deepseek-v4-pro", "name": "DeepSeek-V4-Pro",
     "maxInputTokens": 1000000, "maxOutputTokens": 128000,
     "supportsToolCall": True, "supportsImages": True,
     "reasoning": {"effort": "high", "summary": "auto"}},
    {"id": "orphan-row", "name": "Not advertised by any agent"},
]

CONFIG_DATA = {
    "models": MODEL_ROWS,
    "agents": [
        {"name": "cli", "models": ["glm-5.3", "deepseek-v4-pro", "default",
                                   "hunyuan-image-alpha"]},
        {"name": "craft", "models": ["glm-5.3"]},
    ],
    "enterpriseId": "",
}

ACCOUNT_DATA = {
    "uid": "04eaee2b", "uin": "330101864639", "nickname": "clipboard",
    "type": "personal", "pluginEnabled": True, "phoneNumber": "130****2189",
}


def envelope(data):
    return {"code": 0, "msg": "OK", "requestId": "r1", "data": data}


class FakeResponse:
    def __init__(self, payload, status=200):
        self._body = json.dumps(payload).encode("utf-8")
        self.status = status

    def read(self):
        return self._body

    def __enter__(self):
        return self

    def __exit__(self, *args):
        return False


class FakeOpener:
    """Capture requests and replay canned envelopes keyed by URL substring."""

    def __init__(self, routes):
        self.routes = routes
        self.calls = []

    def open(self, request, timeout=None):
        url = request.full_url
        self.calls.append({"url": url, "headers": {k.lower(): v
                                                   for k, v in request.headers.items()}})
        for needle, payload in self.routes.items():
            if needle in url:
                return FakeResponse(payload)
        raise AssertionError(f"unexpected url {url}")


def install_opener(routes):
    fake = FakeOpener(routes)
    patcher = patch("urllib.request.build_opener", return_value=fake)
    patcher.start()
    return fake, patcher


class EndpointTests(unittest.TestCase):
    def test_normalize_endpoint_strips_gateway_suffix(self):
        # Credentials store base_url with /v2; /v3/config needs the origin only.
        self.assertEqual(cn.normalize_endpoint(f"{CN_ENDPOINT}/v2"), CN_ENDPOINT)
        self.assertEqual(cn.normalize_endpoint(f"{CN_ENDPOINT}/v2/"), CN_ENDPOINT)
        self.assertEqual(cn.normalize_endpoint(CN_ENDPOINT), CN_ENDPOINT)
        self.assertEqual(intl.normalize_endpoint(f"{INTL_ENDPOINT}/v2"), INTL_ENDPOINT)

    def test_normalize_endpoint_rejects_bad_input(self):
        for bad in ("  ", "http://copilot.tencent.com", "https://"):
            with self.assertRaises(cn.WorkBuddyError):
                cn.normalize_endpoint(bad)

    def test_validate_endpoint_allowlist(self):
        cn.validate_endpoint(CN_ENDPOINT, allow_custom=False)
        intl.validate_endpoint(INTL_ENDPOINT, allow_custom=False)
        # The editions keep separate allowlists: a CN host is not international.
        # Each module raises its own WorkBuddyError class, so catch per module.
        with self.assertRaises(intl.WorkBuddyError):
            intl.validate_endpoint(CN_ENDPOINT, allow_custom=False)
        with self.assertRaises(cn.WorkBuddyError):
            cn.validate_endpoint(INTL_ENDPOINT, allow_custom=False)
        cn.validate_endpoint(INTL_ENDPOINT, allow_custom=True)

    def test_endpoint_hosts_match_the_reverse_engineered_origins(self):
        self.assertEqual(cn.PROD_ENDPOINT, CN_ENDPOINT)
        self.assertEqual(intl.PROD_ENDPOINT, INTL_ENDPOINT)
        self.assertIn("wb.tencentbuddy.com", cn.ALLOWED_ENDPOINT_HOSTS)
        self.assertIn("staging-codebuddy.tencent.com", intl.ALLOWED_ENDPOINT_HOSTS)


class UserAgentTests(unittest.TestCase):
    def test_product_line_selects_the_ua_template(self):
        self.assertEqual(cn.build_user_agent("cli", None),
                         "CLI/2.63.2 CodeBuddy/2.63.2")
        self.assertEqual(cn.build_user_agent("ide", None), "WorkBuddy/5.3.14")
        self.assertEqual(cn.build_user_agent("cli", "2.182.0"),
                         "CLI/2.182.0 CodeBuddy/2.182.0")

    def test_cli_ua_matches_the_go_client(self):
        # internal/auth/codebuddycn.UserAgent must stay identical.
        self.assertEqual(cn.build_user_agent("cli", None),
                         "CLI/2.63.2 CodeBuddy/2.63.2")


class FetchTests(unittest.TestCase):
    def test_config_sends_bearer_and_ua_headers(self):
        fake, patcher = install_opener({"/v3/config": envelope(CONFIG_DATA)})
        try:
            result = cn.fetch_config(CN_ENDPOINT, TOKEN,
                                     cn.build_user_agent("cli", None), [])
        finally:
            patcher.stop()
        headers = fake.calls[0]["headers"]
        self.assertEqual(headers["authorization"], f"Bearer {TOKEN}")
        self.assertEqual(headers["user-agent"], "CLI/2.63.2 CodeBuddy/2.63.2")
        self.assertEqual(headers["x-domain"], "copilot.tencent.com")
        self.assertTrue(fake.calls[0]["url"].startswith(f"{CN_ENDPOINT}/v3/config"))
        self.assertEqual(len(result["models"]), len(MODEL_ROWS))

    def test_null_catalog_is_reported_as_invalid_credential(self):
        # The gateway answers HTTP 200 with models == null for anonymous or
        # revoked credentials; that must never look like an empty catalog.
        _, patcher = install_opener({"/v3/config": envelope({"models": None})})
        try:
            with self.assertRaises(cn.WorkBuddyError) as ctx:
                cn.fetch_config(CN_ENDPOINT, TOKEN, "CLI/1.0 CodeBuddy/1.0", [])
        finally:
            patcher.stop()
        self.assertIn("null", str(ctx.exception))

    def test_ua_error_code_is_surfaced(self):
        _, patcher = install_opener({
            "/v3/config": {"code": cn.UA_ERROR_CODE,
                           "msg": "check ua, get coding copilot version error"},
        })
        try:
            with self.assertRaises(cn.WorkBuddyError) as ctx:
                cn.fetch_config(CN_ENDPOINT, TOKEN, "", [])
        finally:
            patcher.stop()
        self.assertIn("12403", str(ctx.exception))

    def test_account_probe_parses_identity(self):
        fake, patcher = install_opener({"/v2/plugin/account": envelope(ACCOUNT_DATA)})
        try:
            account = cn.check_account(CN_ENDPOINT, TOKEN, "CLI/1.0 CodeBuddy/1.0")
        finally:
            patcher.stop()
        self.assertEqual(account["uid"], "04eaee2b")
        self.assertEqual(account["type"], "personal")
        self.assertTrue(account["plugin_enabled"])
        self.assertEqual(account["enterprise_id"], "")  # personal: absent
        self.assertIn("/v2/plugin/account", fake.calls[0]["url"])

    def test_enterprise_catalog_flattens_nested_data(self):
        _, patcher = install_opener({
            "/console/enterprises/e1/config/models":
                {"code": 0, "data": {"data": [{"id": "m1"}, {"id": ""}, "junk"]}},
        })
        try:
            rows = cn.fetch_enterprise_models(CN_ENDPOINT, TOKEN, "CLI/1 CodeBuddy/1", "e1")
        finally:
            patcher.stop()
        self.assertEqual([row["id"] for row in rows], ["m1"])

    def test_http_401_is_reported_as_auth_rejection(self):
        def reject(request, timeout=None):
            raise urllib.error.HTTPError(request.full_url, 401, "Unauthorized", {}, None)

        fake = type("Opener", (), {"open": staticmethod(reject)})()
        patcher = patch("urllib.request.build_opener", return_value=fake)
        patcher.start()
        try:
            with self.assertRaises(cn.WorkBuddyError) as ctx:
                cn.check_account(CN_ENDPOINT, TOKEN, "CLI/1 CodeBuddy/1")
        finally:
            patcher.stop()
        self.assertIn("401", str(ctx.exception))


class ScopeTests(unittest.TestCase):
    def setUp(self):
        self.result = {"data": CONFIG_DATA, "models": MODEL_ROWS}

    def test_cli_scope_follows_agent_order_not_catalog_order(self):
        selected = cn.select_rows(self.result, "cli")
        self.assertEqual([row["id"] for row in selected],
                         ["glm-5.3", "deepseek-v4-pro", "default",
                          "hunyuan-image-alpha"])

    def test_all_scope_returns_the_whole_catalog(self):
        self.assertEqual(len(cn.select_rows(self.result, "all")), len(MODEL_ROWS))

    def test_named_agent_scope(self):
        self.assertEqual([row["id"] for row in cn.select_rows(self.result, "agent=craft")],
                         ["glm-5.3"])

    def test_unknown_agent_names_the_available_agents(self):
        with self.assertRaises(cn.WorkBuddyError) as ctx:
            cn.select_rows(self.result, "agent=nope")
        self.assertIn("cli", str(ctx.exception))

    def test_agent_referencing_an_unlisted_model_raises(self):
        broken = {"data": {"agents": [{"name": "cli", "models": ["ghost"]}]},
                  "models": MODEL_ROWS}
        with self.assertRaises(cn.WorkBuddyError) as ctx:
            cn.select_rows(broken, "cli")
        self.assertIn("ghost", str(ctx.exception))

    def test_agent_models_accept_object_form(self):
        data = {"agents": [{"name": "cli", "models": [{"id": "glm-5.3"}]}]}
        self.assertEqual(cn.agent_ids(data, "cli"), ["glm-5.3"])


class RegistryCatalogTests(unittest.TestCase):
    def test_generation_rows_are_dropped_and_shape_matches_embedded(self):
        entries = cn.to_registry_catalog(cn.select_rows(
            {"data": CONFIG_DATA, "models": MODEL_ROWS}, "cli"))["codebuddy-cn"]
        self.assertEqual([e["id"] for e in entries],
                         ["glm-5.3", "deepseek-v4-pro", "default"])
        first = entries[0]
        self.assertEqual(first["object"], "model")
        self.assertEqual(first["created"], cn.REGISTRY_CREATED)
        self.assertEqual(first["owned_by"], "codebuddy")
        self.assertEqual(first["type"], "codebuddy-cn")
        self.assertEqual(first["display_name"], "GLM-5.3")
        self.assertEqual(first["description"], "GLM-5.3 via CodeBuddy CN")
        self.assertEqual(first["context_length"], 1000000)
        self.assertEqual(first["max_completion_tokens"], 64000)
        self.assertEqual(first["supportedInputModalities"], ["text", "image"])
        self.assertEqual(first["supportedOutputModalities"], ["text"])

    def test_thinking_maps_efforts_and_can_disable_flag(self):
        entries = {e["id"]: e for e in
                   cn.to_registry_catalog(MODEL_ROWS)["codebuddy-cn"]}
        self.assertEqual(entries["glm-5.3"]["thinking"],
                         {"zero_allowed": False, "levels": ["low", "high", "max"]})
        # The single-effort form still yields a level list.
        self.assertEqual(entries["deepseek-v4-pro"]["thinking"],
                         {"zero_allowed": False, "levels": ["high"]})
        # No reasoning block at all: omit thinking rather than invent one.
        self.assertNotIn("thinking", entries["default"])

    def test_duplicate_ids_are_collapsed(self):
        rows = [{"id": "a", "name": "A"}, {"id": "a", "name": "A again"}]
        entries = cn.to_registry_catalog(rows)["codebuddy-cn"]
        self.assertEqual([e["id"] for e in entries], ["a"])

    def test_intl_section_key_and_description_are_the_ai_variant(self):
        entries = intl.to_registry_catalog([{"id": "kimi-k3", "name": "Kimi-K3"}])
        self.assertIn("codebuddy-ai", entries)
        self.assertEqual(entries["codebuddy-ai"][0]["type"], "codebuddy-ai")
        self.assertEqual(entries["codebuddy-ai"][0]["description"],
                         "Kimi-K3 via CodeBuddy AI")
        # The CN writer must not touch the international section and vice versa.
        self.assertNotEqual(cn.REGISTRY_KEY, intl.REGISTRY_KEY)


class OutFileTests(unittest.TestCase):
    def routes(self, rows):
        # Advertise the cli agent set restricted to the rows under test, so each
        # case controls exactly what the catalog offers.
        present = {str(row.get("id")) for row in rows}
        cli = [model_id for model_id in cn.agent_ids(CONFIG_DATA, cn.CLI_AGENT_NAME)
               if model_id in present]
        return {
            "/v3/config": envelope({"models": rows,
                                    "agents": [{"name": "cli", "models": cli}]}),
        }

    def test_merges_only_the_edition_key_preserving_other_sections(self):
        routes = self.routes(MODEL_ROWS)
        _, patcher = install_opener(routes)
        try:
            with tempfile.TemporaryDirectory() as tmp:
                out = Path(tmp) / "nested" / "models.json"
                with patch("sys.stdout", io.StringIO()):
                    rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                                  "--out-file", str(out)])
                self.assertEqual(rc, 0)
                written = json.loads(out.read_text(encoding="utf-8"))
                self.assertEqual(list(written.keys()), ["codebuddy-cn"])
                self.assertEqual([e["id"] for e in written["codebuddy-cn"]],
                                 ["glm-5.3", "deepseek-v4-pro", "default"])
                # Existing shared catalog: only codebuddy-cn is replaced.
                out.write_text(json.dumps(
                    {"claude": [{"id": "claude-opus-5-5"}],
                     "codebuddy-ai": [{"id": "stale-ai"}],
                     "codebuddy-cn": [{"id": "stale"}]}), encoding="utf-8")
                with patch("sys.stdout", io.StringIO()):
                    rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                                  "--out-file", str(out)])
                self.assertEqual(rc, 0)
                written = json.loads(out.read_text(encoding="utf-8"))
                self.assertEqual([e["id"] for e in written["codebuddy-cn"]],
                                 ["glm-5.3", "deepseek-v4-pro", "default"])
                self.assertEqual(written["claude"], [{"id": "claude-opus-5-5"}])
                self.assertEqual(written["codebuddy-ai"], [{"id": "stale-ai"}])
        finally:
            patcher.stop()

    def test_refuses_non_object_json(self):
        _, patcher = install_opener(self.routes(MODEL_ROWS))
        try:
            with tempfile.TemporaryDirectory() as tmp:
                out = Path(tmp) / "models.json"
                out.write_text("[1, 2, 3]", encoding="utf-8")
                with patch("sys.stdout", io.StringIO()):
                    rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                                  "--out-file", str(out)])
                self.assertEqual(rc, 1)
                self.assertEqual(out.read_text(encoding="utf-8"), "[1, 2, 3]")
        finally:
            patcher.stop()

    def test_refuses_when_only_generation_models_are_advertised(self):
        rows = [{"id": "hunyuan-image-alpha", "name": "Hunyuan Image",
                 "tags": ["text-to-image"], "supportsToolCall": False}]
        _, patcher = install_opener(self.routes(rows))
        try:
            with tempfile.TemporaryDirectory() as tmp:
                out = Path(tmp) / "models.json"
                out.write_text('{"codebuddy-cn": [{"id": "keep"}]}', encoding="utf-8")
                with patch("sys.stdout", io.StringIO()):
                    rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                                  "--out-file", str(out)])
                self.assertEqual(rc, 1)
                # A generation-only catalog must not erase the embedded fallback.
                self.assertEqual(json.loads(out.read_text(encoding="utf-8")),
                                 {"codebuddy-cn": [{"id": "keep"}]})
        finally:
            patcher.stop()

    def test_out_file_requires_the_default_cli_scope(self):
        _, patcher = install_opener(self.routes(MODEL_ROWS))
        try:
            with tempfile.TemporaryDirectory() as tmp:
                out = io.StringIO()
                with patch("sys.stderr", out), self.assertRaises(SystemExit):
                    cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                             "--scope", "all", "--out-file", str(Path(tmp) / "m.json")])
                self.assertIn("cli scope", out.getvalue())
        finally:
            patcher.stop()


class CliTests(unittest.TestCase):
    def test_credential_type_guard(self):
        routes = {"/v3/config": envelope(CONFIG_DATA)}
        _, patcher = install_opener(routes)
        try:
            with tempfile.TemporaryDirectory() as tmp:
                good = Path(tmp) / "codebuddy-cn-a.json"
                good.write_text(json.dumps({"type": "codebuddy-cn",
                                            "access_token": TOKEN,
                                            "base_url": f"{CN_ENDPOINT}/v2"}),
                                encoding="utf-8")
                other = Path(tmp) / "codebuddy-ai-b.json"
                other.write_text(json.dumps({"type": "codebuddy-ai",
                                             "access_token": TOKEN}), encoding="utf-8")
                out = io.StringIO()
                with patch("sys.stdout", out):
                    rc = cn.main(["models", "--auth-dir", tmp])
                payload = out.getvalue()
                self.assertEqual(rc, 1)  # one good, one wrong edition
                self.assertIn("glm-5.3", payload)
                self.assertIn("not codebuddy-cn", payload)
        finally:
            patcher.stop()

    def test_credential_base_url_selects_the_edition_endpoint(self):
        fake, patcher = install_opener({"/v3/config": envelope(CONFIG_DATA)})
        try:
            with tempfile.TemporaryDirectory() as tmp:
                cred = Path(tmp) / "codebuddy-ai-x.json"
                cred.write_text(json.dumps({"type": "codebuddy-ai", "access_token": TOKEN,
                                            "base_url": f"{INTL_ENDPOINT}/v2"}),
                                encoding="utf-8")
                with patch("sys.stdout", io.StringIO()):
                    rc = intl.main(["models", "--auth-file", str(cred)])
        finally:
            patcher.stop()
        self.assertEqual(rc, 0)
        self.assertTrue(fake.calls[0]["url"].startswith(INTL_ENDPOINT))
        self.assertEqual(fake.calls[0]["headers"]["x-domain"], "www.codebuddy.ai")

    def test_check_action_reports_personal_account_without_enterprise(self):
        _, patcher = install_opener({"/v2/plugin/account": envelope(ACCOUNT_DATA)})
        try:
            out = io.StringIO()
            with patch("sys.stdout", out):
                rc = cn.main(["check", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                              "--json"])
        finally:
            patcher.stop()
        self.assertEqual(rc, 0)
        account = json.loads(out.getvalue())["accounts"][0]["result"]["account"]
        self.assertEqual(account["type"], "personal")
        self.assertEqual(account["enterprise_id"], "")

    def test_enterprise_action_fails_closed_for_personal_credentials(self):
        _, patcher = install_opener({"/v2/plugin/account": envelope(ACCOUNT_DATA)})
        try:
            out = io.StringIO()
            with patch("sys.stdout", out):
                rc = cn.main(["enterprise", "--token", TOKEN, "--endpoint", CN_ENDPOINT])
        finally:
            patcher.stop()
        self.assertEqual(rc, 1)
        self.assertIn("personal account", out.getvalue())

    def test_ide_product_uses_the_workbuddy_ua(self):
        fake, patcher = install_opener({"/v3/config": envelope(CONFIG_DATA)})
        try:
            with patch("sys.stdout", io.StringIO()):
                rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                              "--product", "ide", "--scope", "agent=craft"])
        finally:
            patcher.stop()
        self.assertEqual(rc, 0)
        self.assertEqual(fake.calls[0]["headers"]["user-agent"], "WorkBuddy/5.3.14")

    def test_json_output_reports_scope_and_count(self):
        _, patcher = install_opener({"/v3/config": envelope(CONFIG_DATA)})
        try:
            out = io.StringIO()
            with patch("sys.stdout", out):
                rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                              "--json"])
        finally:
            patcher.stop()
        self.assertEqual(rc, 0)
        result = json.loads(out.getvalue())["accounts"][0]["result"]
        self.assertEqual(result["scope"], "cli")
        self.assertEqual(result["count"], 4)
        self.assertEqual(result["user_agent"], "CLI/2.63.2 CodeBuddy/2.63.2")

    def test_raw_flag_adds_the_unfiltered_catalog(self):
        _, patcher = install_opener({"/v3/config": envelope(CONFIG_DATA)})
        try:
            out = io.StringIO()
            with patch("sys.stdout", out):
                rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT,
                              "--json", "--raw"])
        finally:
            patcher.stop()
        self.assertEqual(rc, 0)
        result = json.loads(out.getvalue())["accounts"][0]["result"]
        # Scope still narrows models; the raw catalog keeps every row.
        self.assertEqual(result["count"], 4)
        self.assertEqual(result["catalog_count"], len(MODEL_ROWS))
        self.assertEqual([row["id"] for row in result["catalog_models"]],
                         [row["id"] for row in MODEL_ROWS])
        self.assertEqual(result["agents"]["cli"],
                         ["glm-5.3", "deepseek-v4-pro", "default", "hunyuan-image-alpha"])
        self.assertEqual(result["agents"]["craft"], ["glm-5.3"])

    def test_no_automatic_retry_on_network_failure(self):
        attempts = []

        def fail(request, timeout=None):
            attempts.append(request.full_url)
            raise urllib.error.URLError("down")

        fake = type("Opener", (), {"open": staticmethod(fail)})()
        patcher = patch("urllib.request.build_opener", return_value=fake)
        patcher.start()
        try:
            out = io.StringIO()
            with patch("sys.stdout", out):
                rc = cn.main(["models", "--token", TOKEN, "--endpoint", CN_ENDPOINT])
        finally:
            patcher.stop()
        self.assertEqual(rc, 1)
        self.assertEqual(len(attempts), 1)  # single attempt, no retry loop


if __name__ == "__main__":
    unittest.main()
