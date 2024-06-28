CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./output/pecker main.go
cp ./output/pecker ./output/kubectl-pecker
