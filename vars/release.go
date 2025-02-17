//go:build release

package vars

// go build -tags release -ldflags "-w -s" cmd/main.go
const debug = false
