.PHONY: build run test docker-up docker-down seed lint clean

build:
	cd backend && go build -o bin/server ./cmd/server
	cd backend && go build -o bin/seed ./cmd/seed

run:
	cd backend && go run ./cmd/server

test:
	cd backend && go test ./... -v -count=1

test-domain:
	cd backend && go test ./internal/domain/... -v -count=1

seed:
	cd backend && go run ./cmd/seed

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

lint:
	cd backend && go vet ./...

clean:
	rm -rf backend/bin/
	docker-compose down -v

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
