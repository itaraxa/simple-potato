build:
	go build -o out/receiver cmd/receiver/*.go
	go build -o out/sender cmd/sender/*.go

run: build
	nc -u -l 8080 > log/nc.log &
	out/receiver --config config/receiverConfig.json &
	out/sender --config config/senderConfig.json

run_with_log: build
	out/receiver --config config/receiverConfig.json >> log/receiver.log 2>> log/receiver.log &
	out/sender --config config/senderConfig.json >> log/sender.log 2>> log/sender.log &

clean:
	go clean
	rm -rf out/*
	rm log/*
	rm -rf test/downloads/*
