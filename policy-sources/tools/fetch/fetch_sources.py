#!/usr/bin/env python3
import os, sys, json, hashlib, time, urllib.request
from pathlib import Path
import yaml

ROOT = Path(__file__).resolve().parents[2]
REGISTRY = ROOT / "registry" / "sources.yaml"
RAW_DIR = ROOT.parent / "raw"
CACHE_DIR = ROOT.parent / "cache"
PROV_DIR = ROOT.parent / "registry" / "provenance"


def sha256_bytes(b: bytes) -> str:
    return "sha256:" + hashlib.sha256(b).hexdigest()


def ensure_dir(p: Path):
    p.mkdir(parents=True, exist_ok=True)


def fetch_url(url: str) -> bytes:
    with urllib.request.urlopen(url) as response:
        return response.read()


def main():
    with open(REGISTRY, "r") as f:
        config = yaml.safe_load(f)
    ensure_dir(RAW_DIR)
    ensure_dir(CACHE_DIR)
    ensure_dir(PROV_DIR)
    provenance = []

    for source in config.get("sources", []):
        for url in source.get("urls", []):
            try:
                data = fetch_url(url)
                digest = sha256_bytes(data)
                timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
                filename = url.split("/")[-1] or "index.html"
                output_path = RAW_DIR / filename
                with open(output_path, "wb") as fh:
                    fh.write(data)
                provenance.append(
                    {
                        "id": source["id"],
                        "url": url,
                        "hash": digest,
                        "ts": timestamp,
                        "jurisdiction": source.get("jurisdiction"),
                        "domains": source.get("domain_tags", []),
                    }
                )
                print(f"[OK] {source['id']} ← {url} ({digest})")
            except Exception as exc:  # noqa: BLE001
                print(f"[ERR] {source['id']} ← {url}: {exc}", file=sys.stderr)

    prov_path = PROV_DIR / "provenance.json"
    with open(prov_path, "w") as fh:
        json.dump(provenance, fh, indent=2)
    print(f"Provenance written to {prov_path}")


if __name__ == "__main__":
    main()
