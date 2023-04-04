GOOS=windows GOARCH=amd64 go build .


## dev
CGO_ENABLED=1 go get github.com/mattn/go-sqlite3

## debug
go install github.com/go-delve/delve/cmd/dlv@latest
