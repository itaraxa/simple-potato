build:
	go build -o out/receiver cmd/receiver/*.go

run: build
	out/receiver --config config/receiverConfig.json

clear:
	go clean
	rm -rf out/receiver
	rm log/*
