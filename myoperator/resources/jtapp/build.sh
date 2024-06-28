set GOOS=linux
set GOARCH=amd64

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o jtapp main.go