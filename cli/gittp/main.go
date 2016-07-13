package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/adamveld12/gittp"
	"github.com/braintree/manners"
)

func main() {

	config := gittp.ServerConfig{}
	addr, err := parseConfiguration(os.Args[1:], &config)

	if err != nil {
		os.Exit(1)
	}

	config.PostReceive = func(h gittp.HookContext, archive []byte) {
		h.Writeln("Shit fuck")
	}

	sv := manners.NewServer()

	sv.Addr = addr

	handle, err := gittp.NewGitServer(config)
	sv.Handler = handle
	if err != nil {
		log.Fatal("could not open dir", config.Path)
	} else {
		go func() {
			fmt.Printf("Listening for git commands @ %v\n", addr)
			if err := sv.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	<-sig
	fmt.Println("Closing down server...")

	if !sv.BlockingClose() {
		fmt.Println("could not close gracefully")
	}
}

func parseConfiguration(args []string, config *gittp.ServerConfig) (addr string, err error) {
	fSet := flag.NewFlagSet("", flag.ContinueOnError)

	var masterOnly, autocreate bool
	fSet.StringVar(&addr, "addr", ":80", "The addr that gittp listens on")
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

	log.SetFlags(log.Lshortfile | log.Ldate)

	return
}
