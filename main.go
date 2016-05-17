package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/chbatey/go-memcache/memcache"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Info("Starting memcache server")
	log.SetLevel(log.DebugLevel)
	m := memcache.New()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)

	go func() {
		sig := <-c
		log.Infof("Shutting down", sig)
		m.Stop()
	}()

	m.Start()
	m.WaitFor()
}
