.PHONY: build test replay-dry-run diagnose docker-build stack-up stack-down

build:
	go build -o build/streamforge ./cmd/streamforge

test:
	go test ./...

replay-dry-run:
	go run ./cmd/streamforge replay --events examples/orders/source_events.jsonl --topic orders --dry-run

diagnose:
	mkdir -p build
	go run ./cmd/streamforge diagnose --expected examples/orders/expected_iceberg.jsonl --actual examples/orders/actual_iceberg.jsonl --late-after 5m --format json --out build/diagnostics-report.json

docker-build:
	docker build -t streamforge:local .

stack-up:
	docker compose -f deploy/docker-compose.yml up -d

stack-down:
	docker compose -f deploy/docker-compose.yml down -v
