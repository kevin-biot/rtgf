#!/usr/bin/env python3
import json
import sys
import time
import hashlib
from pathlib import Path

TEMPLATE = {
  "@type": "policy:Snapshot",
  "jurisdiction": "",
  "domain": "",
  "effective_date": "",
  "expires_at": "",
  "normative_references": [],
  "controls": {},
  "duties": {},
  "prohibitions": [],
  "policy_snapshot_hash": ""
}

def sha256(data: bytes) -> str:
    return "sha256:" + hashlib.sha256(data).hexdigest()


def main():
    if len(sys.argv) < 5:
        print("usage: normalize_snapshot.py <jurisdiction> <domain> <ref_uri> <out.json>")
        sys.exit(2)

    jurisdiction, domain, ref_uri, output_path = sys.argv[1:5]
    now = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    snapshot = TEMPLATE.copy()
    snapshot["jurisdiction"] = jurisdiction
    snapshot["domain"] = domain
    snapshot["effective_date"] = now
    snapshot["normative_references"] = [ref_uri]

    base = {k: v for k, v in snapshot.items() if k != "policy_snapshot_hash"}
    digest = sha256(json.dumps(base, separators=(",", ":")).encode("utf-8"))
    snapshot["policy_snapshot_hash"] = digest

    Path(output_path).write_text(json.dumps(snapshot, indent=2))
    print(output_path)


if __name__ == "__main__":
    main()
