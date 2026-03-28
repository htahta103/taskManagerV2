.PHONY: api-run api-test web-install web-dev web-build cli-ci ci

api-run:
	go run ./cmd/api

api-test:
	go test ./...
	go vet ./...

web-install:
	cd web && npm ci

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm ci && npm run build

cli-ci:
	cd cli && npm ci && npm run typecheck && npm run build

ci: api-test cli-ci web-build
