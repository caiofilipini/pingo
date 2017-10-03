build:
	go build -o go-ping

.PHONY: run
run: build
	sudo ./go-ping $(host)
