from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path

from .proposals import build_proposals
from .publisher import emit_proposals


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="streamforge-datahub")
    subcommands = parser.add_subparsers(dest="command", required=True)

    publish = subcommands.add_parser("publish", help="build and publish DataHub MCPs")
    publish.add_argument("--spec", required=True, help="dataset metadata spec JSON")
    publish.add_argument("--server", default=os.getenv("DATAHUB_GMS_URL", ""), help="DataHub GMS URL")
    publish.add_argument("--token-env", default="DATAHUB_TOKEN", help="environment variable containing a DataHub token")
    publish.add_argument("--dry-run", action="store_true", help="write proposals without sending them")
    publish.add_argument("--out", default="", help="write proposal JSONL to this path")
    publish.add_argument("--timeout", type=float, default=15.0, help="HTTP timeout in seconds")

    args = parser.parse_args(argv)
    if args.command == "publish":
        return publish_metadata(args)
    parser.error(f"unsupported command {args.command}")
    return 2


def publish_metadata(args: argparse.Namespace) -> int:
    spec_path = Path(args.spec)
    with spec_path.open("r", encoding="utf-8") as handle:
        spec = json.load(handle)

    proposals = build_proposals(spec)
    if args.out:
        write_jsonl(Path(args.out), proposals)
    elif args.dry_run:
        for proposal in proposals:
            print(json.dumps(proposal, sort_keys=True))

    if args.dry_run:
        print(f"prepared {len(proposals)} DataHub proposals", file=sys.stderr)
        return 0

    if not args.server:
        print("publish requires --server or DATAHUB_GMS_URL", file=sys.stderr)
        return 2

    token = os.getenv(args.token_env)
    emit_proposals(args.server, proposals, token=token, timeout=args.timeout)
    print(f"published {len(proposals)} DataHub proposals")
    return 0


def write_jsonl(path: Path, proposals: list[dict]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as handle:
        for proposal in proposals:
            handle.write(json.dumps(proposal, sort_keys=True))
            handle.write("\n")
