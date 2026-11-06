"""Offline regression tests; never call the live service."""
import io
import json
import sys
import tempfile
import unittest
import urllib.error
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent))

from loomy_models import (
    AUTH_ERROR_CODE,
    PROD_BASE,
    TEST_BASE,
    LoomyError,
    build_url,
    fetch_models,
    fetch_quota,
    main,
    normalize_base,
    to_registry_catalog,
    validate_base,
)

TOKEN = "0123456789abcdef0123456789abcdef"

MODEL_ITEMS = [
    {"id": "spark-x", "name": "SparkX", "protocol": "openai", "context_length": 1048576,
     "reasoning": True, "capabilities": {"function_calling": True,
                                         "input_modalities": ["text", "image"]},
     "is_default_model": True},
    {"id": "claude-sonnet", "protocol": "anthropic",
     "limit": {"context": 200000, "output": 8192}},
    {"id": "", "name": "dropped"},
    "plain-string-id",
]

POINTS_ENVELOPE = {
    "code": "000000",
    "desc": "成功",
    "data": {
        "balance": 5000,
        "dailyBalance": 4998,
        "list": [{"chatId": "chat-1", "points": -3}],
        "total": 1,
    },
}


class FakeResponse:
    def __init__(self, payload):
        self._body = json.dumps(payload).encode("utf-8")

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
        self.calls.append({"url": url, "headers": dict(request.headers)})
        for needle, payload in self.routes.items():
            if needle in url:
                return FakeResponse(payload)
        raise AssertionError(f"unexpected url {url}")


def install_opener(routes):
    fake = FakeOpener(routes)
    patcher = patch("urllib.request.build_opener", return_value=fake)
    patcher.start()
    return fake, patcher


class BaseTests(unittest.TestCase):
    def test_normalize_base_appends_api_v1(self):
        self.assertEqual(normalize_base("https://loomyad.xunfei.cn/"), PROD_BASE)
        self.assertEqual(normalize_base(TEST_BASE), TEST_BASE)

    def test_normalize_base_rejects_empty(self):
        with self.assertRaises(LoomyError):
            normalize_base("  ")

    def test_validate_base_allowlist(self):
        validate_base(PROD_BASE, allow_custom=False)
        with self.assertRaises(LoomyError):
            validate_base("https://evil.example.com/api/v1", allow_custom=False)
        validate_base("https://evil.example.com/api/v1", allow_custom=True)
        with self.assertRaises(LoomyError):
            validate_base("http://loomyad.xunfei.cn/api/v1", allow_custom=True)

    def test_build_url_query(self):
        url = build_url(PROD_BASE, "/points/records",
                        {"pageNo": 1, "pageSize": 1, "recordType": "all", "skip": None})
        self.assertEqual(
            url,
            PROD_BASE + "/points/records?pageNo=1&pageSize=1&recordType=all",
        )


class FetchTests(unittest.TestCase):
    def test_models_uses_lowercase_token_header_only(self):
        fake, patcher = install_opener({
            "/models": {"code": "000000", "data": MODEL_ITEMS},
        })
        try:
            models = fetch_models(PROD_BASE, TOKEN)
        finally:
            patcher.stop()
        headers = {k.lower(): v for k, v in fake.calls[0]["headers"].items()}
        self.assertEqual(headers.get("token"), TOKEN)
        self.assertNotIn("authorization", headers)
        self.assertEqual(len(models), 2)  # empty-id dict and non-dict entries dropped
        self.assertEqual([m["id"] for m in models], ["spark-x", "claude-sonnet"])

    def test_models_accepts_paged_data_object(self):
        fake, patcher = install_opener({
            "/models": {"code": "000000", "data": {"records": [{"id": "m1"}]}},
        })
        try:
            models = fetch_models(PROD_BASE, TOKEN)
        finally:
            patcher.stop()
        self.assertEqual([m["id"] for m in models], ["m1"])

    def test_quota_parses_balance_and_optional_window(self):
        fake, patcher = install_opener({
            "/points/records": POINTS_ENVELOPE,
        })
        try:
            quota = fetch_quota(PROD_BASE, TOKEN)
        finally:
            patcher.stop()
        url = fake.calls[0]["url"]
        self.assertIn("pageNo=1", url)
        self.assertIn("pageSize=1", url)
        self.assertIn("recordType=all", url)
        self.assertEqual(quota["balance"], 5000)
        self.assertEqual(quota["daily_balance"], 4998)
        # Missing daily window fields stay None (unknown), never 0.
        self.assertIsNone(quota["daily_limit_points"])
        self.assertIsNone(quota["daily_consumed_points"])

    def test_auth_error_code_raises_relogin_message(self):
        _, patcher = install_opener({
            "/points/records": {"code": AUTH_ERROR_CODE, "desc": "未登录"},
        })
        try:
            with self.assertRaises(LoomyError) as ctx:
                fetch_quota(PROD_BASE, TOKEN)
        finally:
            patcher.stop()
        self.assertIn("100002", str(ctx.exception))

    def test_business_error_code_raises(self):
        _, patcher = install_opener({
            "/models": {"code": "100001", "desc": "参数错误"},
        })
        try:
            with self.assertRaises(LoomyError):
                fetch_models(PROD_BASE, TOKEN)
        finally:
            patcher.stop()


class CliTests(unittest.TestCase):
    def test_credential_type_guard_and_token_resolution(self):
        routes = {"/models": {"code": "000000", "data": [{"id": "spark-x"}]}}
        fake, patcher = install_opener(routes)
        try:
            with tempfile.TemporaryDirectory() as tmp:
                good = Path(tmp) / "loomy-a.json"
                good.write_text(json.dumps(
                    {"type": "loomy", "access_token": TOKEN, "base_url": PROD_BASE}),
                    encoding="utf-8")
                other = Path(tmp) / "other.json"
                other.write_text(json.dumps({"type": "gemini", "access_token": "x"}),
                                 encoding="utf-8")
                out = io.StringIO()
                with patch("sys.stdout", out):
                    rc = main(["models", "--auth-dir", tmp])
                payload = out.getvalue()
                self.assertEqual(rc, 1)  # one good, one rejected
                self.assertIn("spark-x", payload)
                self.assertIn("not loomy", payload)
        finally:
            patcher.stop()

    def test_manual_token_and_json_output(self):
        routes = {"/points/records": POINTS_ENVELOPE}
        _, patcher = install_opener(routes)
        try:
            out = io.StringIO()
            with patch("sys.stdout", out):
                rc = main(["quota", "--token", TOKEN, "--json"])
            payload = json.loads(out.getvalue())
            self.assertEqual(rc, 0)
            quota = payload["accounts"][0]["result"]["quota"]
            self.assertEqual(quota["balance"], 5000)
            self.assertEqual(quota["daily_balance"], 4998)
        finally:
            patcher.stop()

    def test_test_env_flag_routes_to_test_gateway(self):
        routes = {"/models": {"code": "000000", "data": [{"id": "m"}]}}
        fake, patcher = install_opener(routes)
        try:
            with patch("sys.stdout", io.StringIO()):
                rc = main(["models", "--token", TOKEN, "--test-env"])
        finally:
            patcher.stop()
        self.assertEqual(rc, 0)
        self.assertTrue(fake.calls[0]["url"].startswith(TEST_BASE))


CHAT_ROWS = [
    {"id": "spark-x", "name": "Spark X", "type": "chat", "context_length": 1048576,
     "reasoning_efforts": ["none", "low", "medium", "high", "xhigh"],
     "capabilities": {"reasoning": True, "function_calling": True,
                      "input_modalities": ["text", "image"],
                      "output_modalities": ["text"]}},
    {"id": "img-model", "type": "image"},          # non-chat: dropped
    {"id": "spark-x", "type": "chat"},             # duplicate id: dropped
    {"id": "plain", "type": "chat"},               # minimal row
]


class RegistryCatalogTests(unittest.TestCase):
    def test_to_registry_catalog_matches_embedded_shape(self):
        catalog = to_registry_catalog(CHAT_ROWS)
        entries = catalog["loomy"]
        self.assertEqual([e["id"] for e in entries], ["spark-x", "plain"])
        first = entries[0]
        self.assertEqual(first["object"], "model")
        self.assertEqual(first["created"], 1759276800)
        self.assertEqual(first["owned_by"], "iflytek")
        self.assertEqual(first["type"], "loomy")
        self.assertEqual(first["display_name"], "Spark X")
        self.assertEqual(first["context_length"], 1048576)
        self.assertEqual(first["thinking"],
                         {"zero_allowed": True,
                          "levels": ["none", "low", "medium", "high", "xhigh"]})
        self.assertEqual(first["supportedInputModalities"], ["text", "image"])
        self.assertEqual(first["supportedOutputModalities"], ["text"])
        # Minimal row: no thinking block, defaults to text-only modalities.
        minimal = entries[1]
        self.assertNotIn("thinking", minimal)
        self.assertNotIn("context_length", minimal)
        self.assertEqual(minimal["supportedInputModalities"], ["text"])

    def test_out_file_merges_loomy_key_preserving_other_sections(self):
        routes = {"/models": {"code": "000000", "data": CHAT_ROWS}}
        _, patcher = install_opener(routes)
        try:
            with tempfile.TemporaryDirectory() as tmp:
                out = Path(tmp) / "nested" / "models.json"
                # Missing file: created with only the loomy section.
                with patch("sys.stdout", io.StringIO()):
                    rc = main(["models", "--token", TOKEN, "--out-file", str(out)])
                self.assertEqual(rc, 0)
                written = json.loads(out.read_text(encoding="utf-8"))
                self.assertEqual(list(written.keys()), ["loomy"])
                self.assertEqual([e["id"] for e in written["loomy"]], ["spark-x", "plain"])
                # Existing shared catalog: only the loomy key is replaced.
                out.write_text(json.dumps(
                    {"claude": [{"id": "claude-opus-5-5"}], "loomy": [{"id": "stale"}]}),
                    encoding="utf-8")
                routes["/models"] = {"code": "000000", "data": [{"id": "new", "type": "chat"}]}
                with patch("sys.stdout", io.StringIO()):
                    rc = main(["models", "--token", TOKEN, "--out-file", str(out)])
                self.assertEqual(rc, 0)
                written = json.loads(out.read_text(encoding="utf-8"))
                self.assertEqual([e["id"] for e in written["loomy"]], ["new"])
                self.assertEqual(written["claude"], [{"id": "claude-opus-5-5"}])
        finally:
            patcher.stop()

    def test_out_file_refuses_non_object_json(self):
        routes = {"/models": {"code": "000000", "data": CHAT_ROWS}}
        _, patcher = install_opener(routes)
        try:
            with tempfile.TemporaryDirectory() as tmp:
                out = Path(tmp) / "models.json"
                out.write_text("[1, 2, 3]", encoding="utf-8")
                with patch("sys.stdout", io.StringIO()):
                    rc = main(["models", "--token", TOKEN, "--out-file", str(out)])
                self.assertEqual(rc, 1)
                self.assertEqual(out.read_text(encoding="utf-8"), "[1, 2, 3]")
        finally:
            patcher.stop()

    def test_out_file_refuses_empty_chat_list(self):
        routes = {"/models": {"code": "000000", "data": [{"id": "img", "type": "image"}]}}
        _, patcher = install_opener(routes)
        try:
            with tempfile.TemporaryDirectory() as tmp:
                out = Path(tmp) / "models.json"
                out.write_text('{"loomy": [{"id": "keep"}]}', encoding="utf-8")
                with patch("sys.stdout", io.StringIO()):
                    rc = main(["models", "--token", TOKEN, "--out-file", str(out)])
                self.assertEqual(rc, 1)
                # Existing file untouched: no chat rows must not erase the catalog.
                self.assertEqual(json.loads(out.read_text(encoding="utf-8")),
                                 {"loomy": [{"id": "keep"}]})
        finally:
            patcher.stop()


if __name__ == "__main__":
    unittest.main()
