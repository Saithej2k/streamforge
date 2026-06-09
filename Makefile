.PHONY: build test test-python replay-dry-run diagnose datahub-dry-run docker-build stack-up stack-down

build:
	go build -o build/streamforge ./cmd/streamforge

test:
	go test ./...
	PYTHONPATH=python python3 -m unittest discover python/tests

test-python:
	PYTHONPATH=python python3 -m unittest discover python/tests

replay-dry-run:
	go run ./cmd/streamforge replay --events examples/orders/source_events.jsonl --topic orders --dry-run

diagnose:
	mkdir -p build
	go run ./cmd/streamforge diagnose --expected examples/orders/expected_iceberg.jsonl --actual examples/orders/actual_iceberg.jsonl --late-after 5m --format json --out build/diagnostics-report.json

datahub-dry-run:
	mkdir -p build
	PYTHONPATH=python python3 -m streamforge_datahub publish --spec metadata/datahub/orders_dataset.json --dry-run --out build/datahub-mcps.jsonl

docker-build:
	docker build -t streamforge:local .

stack-up:
	docker compose -f deploy/docker-compose.yml up -d

stack-down:
	docker compose -f deploy/docker-compose.yml down -v
