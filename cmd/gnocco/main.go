// Gnocco is a little cache of goodness
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"darvaza.org/darvaza/server/gnocco"
	"darvaza.org/darvaza/shared/version"
)

func main() {
	var confFile string
	var vrs bool
	flag.StringVar(&confFile, "f", "",
		"specify the config file, if empty will try gnocco.conf and /etc/gnocco/gnocco.conf.")
	flag.BoolVar(&vrs, "v", false, "program version")
	flag.Parse()

	if vrs {
		fmt.Fprintf(os.Stdout, "Gnocco version %s, build date %s\n",
			version.Version, version.BuildDate)
		os.Exit(0)
	}

	cf, err := gnocco.NewFromTOML(confFile)
	if err != nil {
		panic(err)
	}

	aserver := gnocco.NewResolver(cf)

	aserver.Run()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)

	for {
		sign := <-signalChan
		switch sign {
		case syscall.SIGTERM:
			aserver.ShutDown()
			cf.Logger().Fatal().Print("Got SIGTERM, stoping as requested")
		case syscall.SIGINT:
			aserver.ShutDown()
			cf.Logger().Fatal().Print("Got SIGINT, stoping as requested")
		case syscall.SIGUSR2:
			cf.Logger().Info().Print("Got SIGUSR2, dumping cache not implemented")
		case syscall.SIGURG:
		default:
			cf.Logger().Warn().Printf("I received %v signal", sign)
		}
	}
}
