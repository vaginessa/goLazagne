GOOS=windows GOARCH=amd64 go build .


## dev
go clean -modcache

## debug
go install github.com/go-delve/delve/cmd/dlv@latest

