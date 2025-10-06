package ciscofonserver

import "github.com/knadh/koanf/v2"

type CiscoFonServer struct {
	config *koanf.Koanf
}

func NewCiscoFonServer() *CiscoFonServer {
	return &CiscoFonServer{}
}

func (s *CiscoFonServer) Run() {
	s.loadConfig()
	go s.startTFTPServer()
	s.startHTTPServer()
}
