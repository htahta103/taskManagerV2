.PHONY: api-run api-test web-install web-dev web-build web-e2e ci ci-full

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

web-e2e:
	cd web && npm ci && npm run build && npx playwright install chromium && npm run test:e2e

ci: api-test web-build

ci-full: api-test web-build web-e2e
