buildAll:
	GOOS=linux GOARCH=amd64 go build -o deployr-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o deployr-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o deployr-macos-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o deployr-macos-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o deployr-windows-amd64.exe main.go

clean:
	rm -f deployr-linux-amd64
	rm -f deployr-linux-arm64
	rm -f deployr-macos-amd64
	rm -f deployr-macos-arm64
	rm -f deployr-windows-amd64.exe


