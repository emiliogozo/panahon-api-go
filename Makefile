-include app.env

createdb:
	docker exec -it pg12 createdb --username=${POSTGRES_ADMIN_USER} --owner=${POSTGRES_DB_USER} ${POSTGRES_DB}

dropdb:
	docker exec -it pg12 dropdb --username=${POSTGRES_ADMIN_USER} ${POSTGRES_DB}

migrateup:
	migrate -path db/migration -database "${DB_SOURCE}" -verbose up

migrateup1:
	migrate -path db/migration -database "${DB_SOURCE}" -verbose up 1

migratedown:
	migrate -path db/migration -database "${DB_SOURCE}" -verbose down

migratedown1:
	migrate -path db/migration -database "${DB_SOURCE}" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate

server:
	go run main.go

mock:
	mockery

swag:
	swag fmt -d main.go,./api
	swag init

test:
	go test -v -cover ./...

short_test:
	go test $(shell go list ./... | grep -v /db/) -cover -short 

build:
	go build -o main main.go


.PHONY: createdb dropdb migrateup migrateup1 migratedown migratedown1 new_migration sqlc server mock swag test short_test build
