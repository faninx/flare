package main

import (
	"github.com/faninx/flare/cmd"
	"github.com/faninx/flare/internal/server"
)

func main() {
	flags := cmd.Parse()
	server.StartDaemon(&flags)
}
