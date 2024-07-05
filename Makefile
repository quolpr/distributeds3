lint:
	golangci-lint run ./...

test:
	go test -v ./...

vet:
	go vet ./...

run:
	docker-compose up --build

