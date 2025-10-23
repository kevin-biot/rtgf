SHELL := /usr/bin/env bash
PY := python3

all: sync normalize-eu-aml

sync:
	$(PY) tools/fetch/fetch_sources.py

normalize-eu-aml:
	$(PY) tools/normalize/normalize_snapshot.py EU payments_aml "CELEX:32018L1673" ../normalized/EU.payments_aml.snapshot.json
	bash tools/sign/sign_snapshot.sh ../normalized/EU.payments_aml.snapshot.json ../normalized/EU.payments_aml.snapshot.signed.json

normalize-ai:
	$(PY) tools/normalize/normalize_snapshot.py EU ai "CELEX:32024R1689" ../normalized/EU.ai.snapshot.json
	bash tools/sign/sign_snapshot.sh ../normalized/EU.ai.snapshot.json ../normalized/EU.ai.snapshot.signed.json
