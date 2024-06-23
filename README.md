# app v1

```bash
# Build for Linux
echo "Building for Linux..."
export GOOS=linux
export GOARCH=amd64
mkdir -p bin/linux
go build -o bin/linux/go_text_to_speech main.go

# Build for Windows
echo "Building for Windows..."
export GOOS=windows
export GOARCH=amd64
mkdir -p bin/windows
go build -o bin/windows/go_text_to_speech.exe main.go

# Build for macOS
echo "Building for macOS..."
export GOOS=darwin
export GOARCH=amd64
mkdir -p bin/macos
go build -o bin/macos/go_text_to_speech main.go

# run in server
$ sudo chmod -R 755 go_text_to_speech
$ sudo ./go_text_to_speech
```
# go_text_to_speech
