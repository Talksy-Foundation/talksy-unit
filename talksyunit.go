package main

import (
	ws "talksy_unit/ws_server"
)

// main starts the server using the custom ws wrapper.
//
// No parameters.
func main() {
	ws.StartServer()
}
