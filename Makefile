tr:
	rm -rf output
	mkdir -p output
	env GOOS=linux GOARCH=arm go build -ldflags "-s -w" -o output/transmission-checker.linux.arm
	env GOOS=darwin GOARCH=amd64 go build -o output/transmission-checker.darwin.amd64

mqttiot:
	rm -f output/mqtt-iot.*
	env GOOS=linux GOARCH=arm go build -ldflags "-s -w" -o output/mqtt-iot.linux.arm ./mqtt-iot
	env GOOS=windows GOARCH=amd64 go build -o output/mqtt-iot.win.x64.exe ./mqtt-iot
	upx -k -o output/mqtt-iot.linux.arm.upx output/mqtt-iot.linux.arm
	ls -lh output/mqtt-iot.*

iptv:
	rm -f output/iptv-m3u8-factory.*
	env GOOS=linux GOARCH=arm go build -ldflags "-s -w" -o output/iptv-m3u8-factory.linux.arm ./iptv-m3u8-factory
	env GOOS=darwin GOARCH=amd64 go build -o output/iptv-m3u8-factory.darwin.amd64 ./iptv-m3u8-factory
