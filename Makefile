.PHONY: build run

build:
	go build -o gession.out cmd/gession/main.go

run:
	go run cmd/gession/main.go

demo:
	rm -f /private/tmp/tmux-501/gession
	vhs demo.tape

prepare-demo-sessions:
	fish prepare-sessions.fish
