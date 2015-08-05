# slog
log for golang
# example
	package main

	import (
		log "github.com/sunvim/slog"
	)

	func main() {

		log.SetLevel(log.DEBUG)
		log.SetRollingDaily("./log", "test.log")
		log.Info("test")

	}