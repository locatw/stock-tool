.PHONY: gen-api lint lint-fix lint-api test

gen-api:
	docker run --rm -v $(PWD)/backend/api:/spec redocly/cli bundle /spec/definition/openapi.yaml --output /spec/definition/bundled.yaml
	cd backend && go generate ./api/gen/...
	rm -f backend/api/definition/bundled.yaml

lint:
	cd backend && go tool golangci-lint run ./...

lint-fix:
	cd backend && go tool golangci-lint run --fix ./...

lint-api:
	docker run --rm -v $(PWD)/backend/api:/spec redocly/cli lint /spec/definition/openapi.yaml --config /spec/config/redocly.yaml

test:
	cd backend && go test ./...
