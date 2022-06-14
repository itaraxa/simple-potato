build:
	go build -o out/receiver cmd/receiver/*.go

run: build
	out/receiver --config config/receiverConfig.json

run_with_log: build
	out/receiver --config config/receiverConfig.json >> log/receiver.log 2>> log/receiver.log

clean:
	go clean
	rm -rf out/receiver
	rm log/*
	rm -rf test/downloads/*
