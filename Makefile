start:
	docker-compose up -d

stop:
	docker-compose down

run:
	go run main.go

database-up:
	docker-compose up -d mysql
