package ciscofonserver

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pin/tftp/v3"
)

// writeHandler is called when client starts file upload to server
func (s *CiscoFonServer) writeHandler(filename string, wt io.WriterTo) error {
	return nil
}

// readHandler is called when client starts file download from server
func (s *CiscoFonServer) readHandler(filename string, rf io.ReaderFrom) error {
	cleanedFilename := cleanPath(filename)
	filepath := filepath.Join(s.config.String("tftp.dir"), cleanedFilename)
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = rf.ReadFrom(file)
	if err != nil {
		return err
	}

	return nil
}

func (s *CiscoFonServer) OnSuccess(stats tftp.TransferStats) {
	s.logRequest("TFTP", "READ", stats.Filename, "OK", stats.RemoteAddr.String())
}

func (s *CiscoFonServer) OnFailure(stats tftp.TransferStats, err error) {
	s.logRequest("TFTP", "READ", stats.Filename, "Err: "+err.Error(), stats.RemoteAddr.String())
}

func (s *CiscoFonServer) startTFTPServer() {
	srv := tftp.NewServer(s.readHandler, s.writeHandler)
	srv.SetHook(s)
	srv.SetTimeout(5 * time.Second)
	log.Printf("Starting TFTP server on port %s, serving files from %s\n", s.config.String("tftp.port"), s.config.String("tftp.dir"))
	err := srv.ListenAndServe(":" + s.config.String("tftp.port"))
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}
}
