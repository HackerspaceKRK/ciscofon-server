package main

import "github.com/HackerspaceKRK/ciscofon-server/pkg/ciscofonserver"

func main() {
	ciscofonserver.NewCiscoFonServer().Run()
}
