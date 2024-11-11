.PHONY: build run

build:
	go build -o gession.out cmd/gession/main.go

run:
	go run cmd/gession/main.go
