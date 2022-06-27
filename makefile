build:
	go build -o out/receiver cmd/receiver/*.go | tee -a log/build.log
	go build -o out/sender cmd/sender/*.go | tee -a log/build.log

buildv2:
	go build -o out/receiver cmd/receiver/*.go | tee -a log/build.log
	go build -o out/senderv2 cmd/sender.v2/*.go | tee -a log/build.log

run: clean build
	out/receiver --config config/receiverConfig.json
	out/sender --config config/senderConfig.json > log/sender.log 2> log/sender.log

runv2: clean buildv2
	out/receiver --config config/receiverConfig.json
	out/senderv2 --config config/senderConfig.json

run_with_log: build
	out/receiver --config config/receiverConfig.json >> log/receiver.log 2>> log/receiver.log &
	out/sender --config config/senderConfig.json >> log/sender.log 2>> log/sender.log &

clean:
	clear
	go clean
	rm -rf out/*
	rm log/*
	rm -rf test/receiver/tmp/*
	rm -rf test/receiver/downloaded/*
	mv test/sender/sended/* test/sender/new/
