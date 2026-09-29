"""Offline regression tests; never call the live service."""
import contextlib
import io
import json
import tempfile
import unittest
import urllib.request
from http.client import HTTPMessage
from pathlib import Path
from unittest.mock import patch

from minimax_checkin import (
    CLAIM_ALREADY_CLAIMED,
    CLAIM_CLAIMED,
    DAY_CLAIMABLE,
    DAY_CLAIMED,
    DAY_UPCOMING,
    ORIGINS,
    CheckinError,
    NoRedirect,
    build_request,
    encode_uri_component,
    main,
    request_json,
    run,
)

CREDENTIAL = {"access_token": "test-only-token", "type": "minimax", "region": "en"}

PANEL_CLAIMABLE = {
    "scene": 2,
    "days": [
        {"day_no": index + 1, "points": (index + 1) * 10, "status": DAY_CLAIMABLE if index == 1 else DAY_UPCOMING,
         "is_today": index == 1}
        for index in range(7)
    ],
}
PANEL_CLAIMED = {
    "scene": 2,
    "days": [
        {"day_no": index + 1, "points": (index + 1) * 10, "status": DAY_CLAIMED if index == 1 else DAY_UPCOMING,
         "is_today": index == 1}
        for index in range(7)
    ],
}
CLAIM = {
    "claim_id": "claim-1", "claim_result": CLAIM_CLAIMED, "day_no": 2, "points": 20,
    "expire_at_ms": 1_800_000_000_000, "panel": PANEL_CLAIMED,
}
IDENTITY = {"base_resp": {"status_code": 0}, "data": {"userInfo": {"realUserID": "user-1"}}}


def ok(data):
    return {"base_resp": {"status_code": 0}, "data": data}


class CheckinTests(unittest.TestCase):
    def test_encode_uri_component_matches_javascript(self):
        # encodeURIComponent percent-encodes reserved characters including '/' and
        # '?', and leaves only the unreserved set unescaped.
        self.assertEqual(encode_uri_component("/a b?c=1"), "%2Fa%20b%3Fc%3D1")
        self.assertEqual(encode_uri_component("A-z_0.9~!*'()"), "A-z_0.9~!*'()")
        self.assertEqual(encode_uri_component("中"), "%E4%B8%AD")

    def test_status_only_does_not_claim(self):
        calls = []

        def transport(request, timeout):
            calls.append(request.method)
            if request.method == "GET" and "/user/info" in request.full_url:
                return IDENTITY
            return ok(PANEL_CLAIMABLE)

        result = run(CREDENTIAL, transport=transport, user_id="user-1")
        self.assertEqual(calls, ["GET"])
        self.assertTrue(result["dryRun"])
        self.assertFalse(result["checked_in_today"])
        self.assertNotIn(CREDENTIAL["access_token"], json.dumps(result))

    def test_claim_when_claimable(self):
        calls = []

        def transport(request, timeout):
            calls.append((request.method, request.full_url))
            if request.method == "GET":
                return ok(PANEL_CLAIMABLE)
            return ok(CLAIM)

        result = run(CREDENTIAL, claim=True, transport=transport, user_id="user-1")
        self.assertTrue(result["claimed"])
        self.assertEqual(result["day_no"], 2)
        self.assertEqual(result["points"], 20)
        self.assertEqual([c[0] for c in calls], ["GET", "POST"])

    def test_already_claimed_today_skips_claim(self):
        calls = []

        def transport(request, timeout):
            calls.append(request.method)
            return ok(PANEL_CLAIMED)

        result = run(CREDENTIAL, claim=True, transport=transport, user_id="user-1")
        self.assertFalse(result["claimed"])
        self.assertEqual(result["reason"], "already-checked-in")
        self.assertEqual(calls, ["GET"])

    def test_unavailable_when_no_claimable_day(self):
        panel = {"scene": 3, "days": [
            {"day_no": i + 1, "points": 10, "status": DAY_UPCOMING, "is_today": i == 1} for i in range(7)
        ]}
        result = run(CREDENTIAL, claim=True, transport=lambda *a: ok(panel), user_id="user-1")
        self.assertFalse(result["claimed"])
        self.assertEqual(result["reason"], "unavailable")

    def test_already_claimed_result_without_prior_status_flag(self):
        # The claim endpoint can report AlreadyClaimed even when status looked claimable.
        claim_already = dict(CLAIM, claim_result=CLAIM_ALREADY_CLAIMED)

        def transport(request, timeout):
            return ok(claim_already) if request.method == "POST" else ok(PANEL_CLAIMABLE)

        result = run(CREDENTIAL, claim=True, transport=transport, user_id="user-1")
        self.assertFalse(result["claimed"])
        self.assertEqual(result["reason"], "already-checked-in")

    def test_resolves_user_id_from_identity_endpoint(self):
        seen = {}

        def transport(request, timeout):
            if "/user/info" in request.full_url:
                seen["identity"] = request.full_url
                return IDENTITY
            seen["status"] = request.full_url
            return ok(PANEL_CLAIMABLE)

        run(CREDENTIAL, transport=transport)
        self.assertIn("user_id=0", seen["identity"])
        self.assertIn("user_id=user-1", seen["status"])

    def test_status_base_resp_error_is_surfaced(self):
        def transport(request, timeout):
            if "/user/info" in request.full_url:
                return IDENTITY
            return {"base_resp": {"status_code": 1001, "status_msg": "not signed in"}}

        with self.assertRaisesRegex(CheckinError, "not signed in"):
            run(CREDENTIAL, transport=transport, user_id="user-1")

    def test_signed_headers_present(self):
        captured = {}

        def transport(request, timeout):
            captured["headers"] = {k.lower(): v for k, v in request.headers.items()}
            return IDENTITY if "/user/info" in request.full_url else ok(PANEL_CLAIMABLE)

        run(CREDENTIAL, transport=transport, user_id="user-1")
        headers = captured["headers"]
        self.assertTrue(headers["authorization"].startswith("Bearer "))
        self.assertEqual(headers["user-agent"], "MiniMaxCode")
        self.assertEqual(len(headers["yy"]), 32)
        self.assertEqual(len(headers["x-signature"]), 32)
        self.assertTrue(headers["x-timestamp"].isdigit())

    def test_signature_is_deterministic_for_fixed_time(self):
        now_ms = 1_800_000_000_000
        first = build_request(
            ORIGINS["en"], "/minimax-cloud/api/v1/signin/status", "GET", "t",
            [("a", "1")], None, now_ms,
        )
        second = build_request(
            ORIGINS["en"], "/minimax-cloud/api/v1/signin/status", "GET", "t",
            [("a", "1")], None, now_ms,
        )
        self.assertEqual(first.headers["Yy"], second.headers["Yy"])
        self.assertEqual(first.headers["X-signature"], second.headers["X-signature"])

    def test_untrusted_origin_rejected(self):
        with self.assertRaises(CheckinError):
            run(CREDENTIAL, origin="https://evil.example", transport=lambda *a: {}, user_id="u")

    def test_invalid_credentials(self):
        for credential in (None, {}, {"access_token": 1}, {"access_token": "  "}):
            with self.assertRaises(CheckinError):
                run(credential, transport=lambda *a: {})

    def test_region_inference(self):
        cn = run({"access_token": "t", "base_url": "https://agent.minimax.cn/mavis/x"},
                 transport=lambda *a: ok(PANEL_CLAIMABLE), user_id="user-1")
        self.assertEqual(cn["region"], "cn")
        en = run({"access_token": "t", "base_url": "https://agent.minimax.io/mavis/x"},
                 transport=lambda *a: ok(PANEL_CLAIMABLE), user_id="user-1")
        self.assertEqual(en["region"], "en")

    def test_redirect_refused(self):
        request = urllib.request.Request("https://example.com")  # noqa: S310 -- fixed HTTPS test URL
        self.assertIsNone(NoRedirect().redirect_request(
            request, io.BytesIO(), 302, "", HTTPMessage(), "https://example.com"))

    def test_redirect_open_is_refused(self):
        # A 302 response must not be followed, so the bearer token is never re-sent.
        request = build_request(ORIGINS["en"], "/minimax-cloud/api/v1/signin/status", "GET",
                                "t", [("a", "1")], None, 1_800_000_000_000)
        with patch("minimax_checkin.urllib.request.build_opener") as factory:
            factory.return_value.open.side_effect = OSError("test-only-token")
            with self.assertRaisesRegex(CheckinError, "^Network request failed"):
                request_json(request)
            # The token must not leak into the sanitized error message.

    @patch("minimax_checkin.run")
    def test_cli_batch_continues_and_filters(self, runner):
        runner.return_value = {"dryRun": False, "region": "en", "user_id": "u", "panel": {"days": []},
                               "checked_in_today": False, "claimed": True}
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "minimax-1.json").write_text("invalid", encoding="utf-8")
            (root / "minimax-2.json").write_text(json.dumps(CREDENTIAL), encoding="utf-8")
            (root / "other.json").write_text(json.dumps(CREDENTIAL), encoding="utf-8")
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                code = main(["--auth-dir", directory, "--prefix", "minimax-", "--json"])
            self.assertEqual(code, 1)
            self.assertEqual(json.loads(output.getvalue())["summary"],
                             {"total": 2, "succeeded": 1, "failed": 1})
            runner.assert_called_once()
            self.assertTrue(runner.call_args.kwargs["claim"])

    @patch("minimax_checkin.run")
    def test_cli_status_and_default_action(self, runner):
        runner.return_value = {"dryRun": True, "region": "en", "user_id": "u", "panel": {"days": []},
                               "checked_in_today": False}
        for action, expected in [([], True), (["status"], False)]:
            with contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(main(action + ["--token", "test-token", "--json"]), 0)
            self.assertEqual(runner.call_args.kwargs["claim"], expected)

    def test_cli_invalid_options(self):
        for args in [["--prefix", "minimax"], ["status", "--claim", "--token", "x"],
                     ["--auth-file", "x", "--auth-dir", "y"], ["--user-id", "u"]]:
            with contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
                main(args)

    def test_rejects_foreign_credential_type(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "claude.json"
            path.write_text(json.dumps({"type": "claude", "access_token": "x"}), encoding="utf-8")
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                code = main(["--auth-file", str(path), "--json"])
            payload = json.loads(output.getvalue())
            self.assertEqual(code, 1)
            self.assertIn("not minimax", payload["accounts"][0]["error"])


if __name__ == "__main__":
    unittest.main()
