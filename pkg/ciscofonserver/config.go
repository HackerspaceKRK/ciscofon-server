package ciscofonserver

import (
	"log"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func (s *CiscoFonServer) loadConfig() {
	s.config = koanf.New(".")
	if err := s.config.Load(file.Provider("ciscofonserver.yaml"), yaml.Parser()); err != nil {
		log.Printf("Failed to load config from ciscofonserver.yaml, trying /config/ciscofonserver.yaml: %v", err)
		if err := s.config.Load(file.Provider("/config/ciscofonserver.yaml"), yaml.Parser()); err != nil {
			log.Fatalf("error loading config from: %v", err)
		}

	}
}
