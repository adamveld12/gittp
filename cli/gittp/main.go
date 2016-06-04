package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adamveld12/gittp"
)

func main() {

	config := gittp.ServerConfig{}
	port, err := parseConfiguration(os.Args[1:], &config)

	if err != nil {
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%s", port)

	handle, err := gittp.NewGitServer(config)

	if err != nil {
		log.Fatal("could not open dir", config.Path)
	} else {
		log.Fatal(http.ListenAndServe(addr, handle))
	}
}

func parseConfiguration(args []string, config *gittp.ServerConfig) (port string, err error) {
	fSet := flag.NewFlagSet("", flag.ContinueOnError)

	var masterOnly, autocreate bool
	fSet.StringVar(&port, "port", "80", "The port that gittp listens on")
	fSet.StringVar(&config.Path, "path", "./repositories", "The path that gittp stores pushed repositories")
	fSet.BoolVar(&masterOnly, "masteronly", false, "Only allow pushing to master")
	fSet.BoolVar(&autocreate, "autocreate", false, "Auto creates repositories if they have not been created")
	fSet.BoolVar(&config.Debug, "debug", false, "Enables debug logging")

	err = fSet.Parse(args)

	if autocreate {
		config.PreCreate = gittp.CreateRepo
	}

	if masterOnly {
		config.PreReceive = gittp.MasterOnly
	}

	return
}
