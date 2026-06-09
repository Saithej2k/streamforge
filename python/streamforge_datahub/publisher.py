from __future__ import annotations

import json
import urllib.error
import urllib.request
from typing import Any


def emit_proposals(server: str, proposals: list[dict[str, Any]], token: str | None = None, timeout: float = 15.0) -> None:
    endpoint = f"{server.rstrip('/')}/aspects?action=ingestProposal"
    for proposal in proposals:
        emit_proposal(endpoint, proposal, token=token, timeout=timeout)


def emit_proposal(endpoint: str, proposal: dict[str, Any], token: str | None = None, timeout: float = 15.0) -> None:
    payload = json.dumps({"proposal": proposal}).encode("utf-8")
    headers = {
        "Content-Type": "application/json",
        "X-RestLi-Protocol-Version": "2.0.0",
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"

    request = urllib.request.Request(endpoint, data=payload, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            response.read()
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"DataHub rejected {proposal['aspectName']} for {proposal['entityUrn']}: {exc.code} {body}") from exc
    except urllib.error.URLError as exc:
        raise RuntimeError(f"DataHub endpoint is unavailable: {exc}") from exc
