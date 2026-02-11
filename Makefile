.PHONY: lint lint-fix test

lint:
	cd backend && go tool golangci-lint run ./...

lint-fix:
	cd backend && go tool golangci-lint run --fix ./...

test:
	cd backend && go test ./...
