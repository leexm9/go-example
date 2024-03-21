package main

import (
	"flag"
	"fmt"
	"go-example/hot_switch"
	"go-example/hot_switch/demo/g"
	"go-example/hot_switch/utils"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var pluginDir, pidFile string

func init() {
	flag.StringVar(&pluginDir, "pluginDir", "", "the directory contains your plugins")
	flag.StringVar(&pidFile, "pidFile", "", "pid file")
}

func main() {
	flag.Parse()

	g.Logg = log.Default()
	absDir, err := filepath.Abs(pluginDir)
	if err != nil {
		panic(err)
	}
	if e := utils.IsDirectory(absDir, ""); e != nil {
		panic(e)
	}
	if pidFile == "" {
		panic("no --pidFile")
	}

	pid := fmt.Sprint(os.Getpid())
	if e := os.WriteFile(pidFile, []byte(pid), 0644); e != nil {
		panic(e)
	}

	g.PluginSwitch = hot_switch.NewPluginSwitch(g.Logg, absDir)
	details, err := g.PluginSwitch.InitLoad()
	if err != nil {
		panic(err)
	} else if len(details) == 0 {
		panic("no module is found in " + absDir)
	} else {
		g.Logg.Printf("<hotswitch> %d modules loaded. details: [%s]", len(details), details)
	}

	go func() {
		heartbeat()
		for range time.Tick(time.Second * 3) {
			heartbeat()
		}
	}()

	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

loop:
	for {
		select {
		case sig := <-chSignal:
			g.Logg.Printf("signal received: %v", sig)
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				break loop
			case syscall.SIGUSR1:
				g.Logg.Printf("<hotswitch> reloading...")
				details, err = g.PluginSwitch.Reload()
				if err != nil {
					panic(err)
				} else if len(details) == 0 {
					panic("no module is found in " + absDir)
				} else {
					g.Logg.Printf("<hotswitch> %d modules loaded. details: [%s]", len(details), details)
				}
				heartbeat()
			}
		}
	}

	signal.Reset(syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	g.Logg.Println("THE END")
}

func heartbeat() {
	rs, err := g.PluginSwitch.Current().Invoke("first.Add", 2, 4)
	if err != nil {
		g.Logg.Printf("Invoke func error: %s", err)
	} else {
		g.Logg.Printf("Invoke func result: %d", rs)
	}
}
