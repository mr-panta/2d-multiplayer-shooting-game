build/client:
	make build/client/macos
	make build/client/windows
	mkdir -p bundle
	rm -rf bundle/*
	echo "client.exe" > bundle/windows_start.bat
	mv bin/* bundle
	cp -R asset bundle

build/client/macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/client cmd/client/*

build/client/windows:
	GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 go build -o bin/client.exe cmd/client/*

run/client:
	GORUN=1 go run cmd/client/*

run/server:
	go run cmd/server/*