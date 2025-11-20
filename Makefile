.PHONY: run stop test lint
run:
	docker-compose up --build

stop:
	docker-compose down

test:
	go test ./...

lint:
	golangci-lint run