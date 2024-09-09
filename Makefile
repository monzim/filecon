build-mac:
	GOOS=darwin GOARCH=amd64 go build -o filecon-macos
	GOOS=darwin GOARCH=arm64 go build -o filecon-macos-arm64
build-linux:
	GOOS=linux GOARCH=amd64 go build -o filecon-linux
	GOOS=linux GOARCH=arm64 go build -o filecon-linux-arm64
build-windows:
	GOOS=windows GOARCH=amd64 go build -o filecon-windows.exe
	GOOS=windows GOARCH=arm64 go build -o filecon-windows-arm64.exe
