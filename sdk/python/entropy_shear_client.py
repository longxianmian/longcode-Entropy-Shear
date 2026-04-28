"""Minimal HTTP client for the Entropy Shear API.

This SDK only wraps urllib. It never evaluates rules locally, caches
results, retries, or implements auth. Designed to drop into a Python 3.9+
project with no third-party dependencies.
"""

from __future__ import annotations

import json
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from typing import Any, Optional


__all__ = ["EntropyShearClient", "EntropyShearError"]


@dataclass
class EntropyShearError(Exception):
    """Raised on non-2xx responses from the Entropy Shear API."""

    status: int
    code: Optional[str] = None
    detail: Optional[str] = None

    def __post_init__(self) -> None:  # pragma: no cover - trivial
        bits = [f"entropy-shear {self.status}"]
        if self.code:
            bits.append(self.code)
        if self.detail:
            bits.append(self.detail)
        super().__init__(": ".join(bits))


class EntropyShearClient:
    """Thin wrapper over POST /shear, GET /ledger/*, GET /health."""

    def __init__(self, base_url: str, *, timeout: float = 10.0) -> None:
        if not base_url:
            raise ValueError("base_url is required")
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    # ---- public API ---------------------------------------------------

    def health(self) -> dict[str, Any]:
        return self._request("GET", "/health")

    def shear(self, *, policy: dict[str, Any], facts: dict[str, Any]) -> dict[str, Any]:
        return self._request("POST", "/shear", body={"policy": policy, "facts": facts})

    def get_ledger_record(self, shear_id: str) -> dict[str, Any]:
        if not shear_id:
            raise ValueError("shear_id is required")
        return self._request("GET", f"/ledger/{urllib.parse.quote(shear_id, safe='')}")

    def verify_ledger(self) -> dict[str, Any]:
        return self._request("GET", "/ledger/verify")

    # ---- internals ----------------------------------------------------

    def _request(
        self,
        method: str,
        path: str,
        *,
        body: Optional[dict[str, Any]] = None,
    ) -> dict[str, Any]:
        url = self.base_url + path
        data: Optional[bytes] = None
        headers = {"Accept": "application/json"}
        if body is not None:
            data = json.dumps(body, ensure_ascii=False).encode("utf-8")
            headers["Content-Type"] = "application/json; charset=utf-8"
        req = urllib.request.Request(url=url, data=data, method=method, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                raw = resp.read()
                return _safe_decode(raw)
        except urllib.error.HTTPError as e:
            raw = e.read() if hasattr(e, "read") else b""
            payload = _safe_decode(raw)
            code = payload.get("error") if isinstance(payload, dict) else None
            detail = payload.get("detail") if isinstance(payload, dict) else (
                raw.decode("utf-8", errors="replace") if raw else None
            )
            raise EntropyShearError(status=e.code, code=code, detail=detail) from None


def _safe_decode(raw: bytes) -> Any:
    if not raw:
        return {}
    try:
        return json.loads(raw.decode("utf-8"))
    except (json.JSONDecodeError, UnicodeDecodeError):
        return {"raw": raw.decode("utf-8", errors="replace")}
