.PHONY: build run test migrate migrate-up migrate-down migrate-create migrate-version

# Build the application
build:
	go build -o bin/app cmd/main.go

# Run the application
run:
	export APP_CONFIG_FILE=/Users/calvin/Documents/working/me/interview/btaskee/home_assignment/real-time-quiz/real-time-quiz-backend/config.yaml && go run cmd/main.go

# Run tests
test:
	go test -v ./...

# Database migration commands
# Usage: make migrate-create name=add_quiz_code
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required. Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations/versioned -seq $(name)

# Run all pending migrations
migrate-up:
	migrate -path migrations/versioned -database "${DB_URL}" up

# Roll back all migrations
migrate-down:
	migrate -path migrations/versioned -database "${DB_URL}" down

# Roll back one migration
migrate-down-one:
	migrate -path migrations/versioned -database "${DB_URL}" down 1

# Show current migration version
migrate-version:
	migrate -path migrations/versioned -database "${DB_URL}" version

# Apply specific migration version
# Usage: make migrate-goto version=2
migrate-goto:
	@if [ -z "$(version)" ]; then \
		echo "Error: version parameter is required. Usage: make migrate-goto version=N"; \
		exit 1; \
	fi
	migrate -path migrations/versioned -database "${DB_URL}" goto $(version)

# Run database migration (alias for migrate-up)
migrate: migrate-up