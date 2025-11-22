.PHONY: run stop test lint test-integration
run:
	docker-compose up --build

stop:
	docker-compose down

test:
	go test ./...

lint:
	golangci-lint run

test-integration:
	go test -v -tags=integration ./tests/...