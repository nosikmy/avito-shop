.PHONY: test
test:
	go test -v ./...

.PHONY: test-integration
test-integration:
	go test -v -tags integration ./...

.PHONY: get-coverage
get-coverage:
	go test -v -tags integration --cover -covermode=count ./... -coverprofile=cover.out
	go tool cover -html=cover.out

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5
	golangci-lint run ./...