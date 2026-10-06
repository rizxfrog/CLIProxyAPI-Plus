"""Offline regression tests for workbuddy_checkin; never call the live service."""
import base64
import contextlib
import io
import json
import tempfile
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch

from workbuddy_checkin import (
    REFRESH_PATH,
    _run_one,
    load_session,
    main,
    refresh_credential,
    save_session,
)

HOST = "https://copilot.tencent.com"


def fake_jwt(sub: str) -> str:
    """Build an unsigned JWT whose payload carries sub, enough for uid decoding."""
    def seg(obj):
        raw = base64.urlsafe_b64encode(json.dumps(obj).encode("utf-8"))
        return raw.rstrip(b"=").decode("ascii")

    return f"{seg({'alg': 'none'})}.{seg({'sub': sub})}.sig"


CLI_CREDENTIAL = {
    "type": "codebuddy-cn",
    "access_token": fake_jwt("uid-1"),
    "refresh_token": "refresh-old",
    "base_url": f"{HOST}/v2",
    "enterprise_id": "ent-1",
}

DESKTOP_CREDENTIAL = {
    "auth": {"accessToken": fake_jwt("uid-2"), "refreshToken": "refresh-old", "domain": "example.test"},
    "account": {"uid": "uid-2", "enterpriseId": "ent-2"},
}

REFRESH_OK = {"status": 200, "payload": {"code": 0, "data": {
    "accessToken": "access-new", "refreshToken": "refresh-new", "expiresIn": 3600}}}
STATUS_OK = {"status": 200, "payload": {"code": 0, "data": {"today_checked_in": False}}}
CLAIM_OK = {"status": 200, "payload": {"code": 0, "data": {}}}


def write_credential(directory: str, name: str, content: dict) -> Path:
    path = Path(directory) / name
    path.write_text(json.dumps(content), encoding="utf-8")
    return path


class LoadSessionTests(unittest.TestCase):
    def test_cli_layout_carries_refresh_token_and_strips_v2(self):
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "codebuddy-cn-1.json", CLI_CREDENTIAL)
            with contextlib.redirect_stdout(io.StringIO()):
                session = load_session(path)
        self.assertEqual(session["token"], CLI_CREDENTIAL["access_token"])
        self.assertEqual(session["uid"], "uid-1")
        self.assertEqual(session["endpoint"], HOST)
        self.assertEqual(session["refresh_token"], "refresh-old")
        self.assertEqual(session["_layout"], "cli")

    def test_desktop_layout_carries_refresh_token(self):
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "auth.info", DESKTOP_CREDENTIAL)
            with contextlib.redirect_stdout(io.StringIO()):
                session = load_session(path)
        self.assertEqual(session["uid"], "uid-2")
        self.assertEqual(session["refresh_token"], "refresh-old")
        self.assertEqual(session["_layout"], "desktop")
        self.assertEqual(session["enterpriseId"], "ent-2")


class RefreshTests(unittest.TestCase):
    def test_request_shape_matches_plugin_refresh_endpoint(self):
        calls = []

        def fake(url, headers, body, timeout):
            calls.append((url, headers, body))
            return REFRESH_OK

        session = {"token": "old", "uid": "u", "domain": "domain.test", "refresh_token": "rt"}
        with patch("workbuddy_checkin._http_json", fake):
            refreshed = refresh_credential(HOST, session, 15)

        self.assertEqual(calls[0][0], f"{HOST}{REFRESH_PATH}")
        url, headers, body = calls[0]
        self.assertEqual(headers["X-Refresh-Token"], "rt")
        self.assertEqual(headers["X-Auth-Refresh-Source"], "plugin")
        self.assertEqual(headers["X-Domain"], "domain.test")
        self.assertNotIn("Authorization", headers)
        self.assertEqual(body, {})
        self.assertEqual(refreshed["token"], "access-new")
        self.assertEqual(refreshed["refresh_token"], "refresh-new")
        self.assertEqual(refreshed["expires_in"], 3600)
        self.assertTrue(refreshed["expired"])

    def test_domain_defaults_to_endpoint_host(self):
        seen = {}

        def fake(url, headers, body, timeout):
            seen.update(headers)
            return REFRESH_OK

        with patch("workbuddy_checkin._http_json", fake):
            refresh_credential("https://www.codebuddy.ai", {"refresh_token": "rt"}, 15)
        self.assertEqual(seen["X-Domain"], "www.codebuddy.ai")

    def test_missing_refresh_token_is_an_error(self):
        with self.assertRaisesRegex(Exception, "refresh_token"):
            refresh_credential(HOST, {"token": "t"}, 15)

    def test_rejected_refresh_reports_code(self):
        with patch("workbuddy_checkin._http_json",
                   lambda *a: {"status": 200, "payload": {"code": 401, "msg": "expired"}}):
            with self.assertRaisesRegex(Exception, "401"):
                refresh_credential(HOST, {"refresh_token": "rt"}, 15)

    def test_response_without_access_token_is_rejected(self):
        with patch("workbuddy_checkin._http_json",
                   lambda *a: {"status": 200, "payload": {"code": 0, "data": {}}}):
            with self.assertRaisesRegex(Exception, "accessToken"):
                refresh_credential(HOST, {"refresh_token": "rt"}, 15)

    def test_refresh_token_is_kept_when_response_omits_it(self):
        response = {"status": 200, "payload": {"code": 0, "data": {"accessToken": "a", "expiresIn": 60}}}
        with patch("workbuddy_checkin._http_json", lambda *a: response):
            refreshed = refresh_credential(HOST, {"refresh_token": "rt"}, 15)
        self.assertEqual(refreshed["refresh_token"], "rt")


class SaveSessionTests(unittest.TestCase):
    def test_cli_layout_is_rotated_in_place(self):
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "c.json", CLI_CREDENTIAL)
            save_session({"refresh_token": "r", "_path": str(path), "_layout": "cli"},
                         {"token": "access-new", "refresh_token": "refresh-new",
                          "expires_in": 3600, "expired": "2030-01-01T00:00:00Z"})
            saved = json.loads(path.read_text(encoding="utf-8"))
        self.assertEqual(saved["access_token"], "access-new")
        self.assertEqual(saved["refresh_token"], "refresh-new")
        self.assertEqual(saved["expired"], "2030-01-01T00:00:00Z")
        # Unrelated fields survive the rewrite.
        self.assertEqual(saved["base_url"], f"{HOST}/v2")
        self.assertEqual(saved["enterprise_id"], "ent-1")
        self.assertIn("timestamp", saved)

    def test_desktop_layout_rotates_only_the_auth_block(self):
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "auth.info", DESKTOP_CREDENTIAL)
            save_session({"_path": str(path), "_layout": "desktop"},
                         {"token": "access-new", "refresh_token": "refresh-new",
                          "expires_in": 60, "expired": "2030-01-01T00:00:00Z"})
            saved = json.loads(path.read_text(encoding="utf-8"))
        self.assertEqual(saved["auth"]["accessToken"], "access-new")
        self.assertEqual(saved["auth"]["refreshToken"], "refresh-new")
        self.assertEqual(saved["auth"]["domain"], "example.test")
        self.assertEqual(saved["account"]["uid"], "uid-2")
        # The desktop schema must not gain CLIProxyAPI-only keys.
        self.assertEqual(sorted(saved), ["account", "auth"])

    def test_no_temp_file_is_left_behind(self):
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "c.json", CLI_CREDENTIAL)
            save_session({"_path": str(path), "_layout": "cli"},
                         {"token": "a", "refresh_token": "r", "expires_in": 0, "expired": ""})
            leftovers = [p.name for p in Path(directory).iterdir() if p.name.endswith(".tmp")]
        self.assertEqual(leftovers, [])


class RunOneTests(unittest.TestCase):
    ARGS = SimpleNamespace(action="checkin", refresh_token=True, timeout=15,
                           endpoint=HOST, staging=False)

    def test_refresh_runs_before_status_and_uses_new_token(self):
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "c.json", CLI_CREDENTIAL)
            calls = []

            def fake(url, headers, body, timeout):
                calls.append((url, headers.get("Authorization")))
                if url.endswith(REFRESH_PATH):
                    return REFRESH_OK
                if "checkin-activity-status" in url:
                    return STATUS_OK
                return CLAIM_OK

            with contextlib.redirect_stdout(io.StringIO()):
                with patch("workbuddy_checkin._http_json", fake):
                    code = _run_one(self.ARGS, load_session(path), HOST)

            self.assertEqual(code, 0)
            self.assertTrue(calls[0][0].endswith(REFRESH_PATH))
            # Check-in requests must carry the freshly rotated token, not the stale one.
            self.assertEqual(calls[1][1], "Bearer access-new")
            self.assertEqual(calls[2][1], "Bearer access-new")
            saved = json.loads(path.read_text(encoding="utf-8"))
            self.assertEqual(saved["access_token"], "access-new")

    def test_refresh_failure_skips_checkin_and_leaves_file_untouched(self):
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "c.json", CLI_CREDENTIAL)
            calls = []

            def fake(url, headers, body, timeout):
                calls.append(url)
                return {"status": 200, "payload": {"code": 401, "msg": "expired"}}

            with contextlib.redirect_stdout(io.StringIO()):
                with patch("workbuddy_checkin._http_json", fake):
                    code = _run_one(self.ARGS, load_session(path), HOST)

            self.assertEqual(code, 1)
            self.assertEqual(len(calls), 1)
            self.assertEqual(json.loads(path.read_text(encoding="utf-8")),
                             json.loads(json.dumps(CLI_CREDENTIAL)))

    def test_status_action_also_refreshes_first(self):
        args = SimpleNamespace(action="status", refresh_token=True, timeout=15)
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "c.json", CLI_CREDENTIAL)
            calls = []

            def fake(url, headers, body, timeout):
                calls.append((url, headers.get("Authorization")))
                return REFRESH_OK if url.endswith(REFRESH_PATH) else STATUS_OK

            with contextlib.redirect_stdout(io.StringIO()):
                with patch("workbuddy_checkin._http_json", fake):
                    code = _run_one(args, load_session(path), HOST)

            self.assertEqual(code, 0)
            self.assertTrue(calls[0][0].endswith(REFRESH_PATH))
            self.assertEqual(calls[1][1], "Bearer access-new")

    def test_no_refresh_flag_keeps_old_token(self):
        args = SimpleNamespace(action="checkin", refresh_token=False, timeout=15)
        with tempfile.TemporaryDirectory() as directory:
            path = write_credential(directory, "c.json", CLI_CREDENTIAL)
            calls = []

            def fake(url, headers, body, timeout):
                calls.append((url, headers.get("Authorization")))
                return STATUS_OK if "checkin-activity-status" in url else CLAIM_OK

            with contextlib.redirect_stdout(io.StringIO()):
                with patch("workbuddy_checkin._http_json", fake):
                    code = _run_one(args, load_session(path), HOST)

            self.assertEqual(code, 0)
            self.assertTrue(all(REFRESH_PATH not in url for url, _ in calls))
            self.assertEqual(calls[0][1], f"Bearer {CLI_CREDENTIAL['access_token']}")


class CliTests(unittest.TestCase):
    def test_refresh_token_conflicts_with_manual_token(self):
        with contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
            main(["--refresh-token", "--token", "t", "--uid", "u"])

    def test_batch_continues_when_one_credential_cannot_refresh(self):
        with tempfile.TemporaryDirectory() as directory:
            write_credential(directory, "codebuddy-1.json", {"access_token": fake_jwt("u1")})
            write_credential(directory, "codebuddy-2.json", CLI_CREDENTIAL)

            def fake(url, headers, body, timeout):
                if url.endswith(REFRESH_PATH):
                    return REFRESH_OK
                if "checkin-activity-status" in url:
                    return STATUS_OK
                return CLAIM_OK

            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                with patch("workbuddy_checkin._http_json", fake):
                    code = main(["--auth-dir", directory, "--prefix", "codebuddy-",
                                 "--refresh-token"])
            self.assertEqual(code, 1)  # first credential has no refresh token
            self.assertIn("凭证刷新失败", output.getvalue())
