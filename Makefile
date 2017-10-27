build:
	go build -o pingo

.PHONY: run
run: build
	sudo ./pingo $(host)
