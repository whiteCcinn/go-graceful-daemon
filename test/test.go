package main

import (
	daemon "github.com/whiteCcinn/go-graceful-daemon"
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	daemon.SetSigHandler(sigtermMyHandler, syscall.SIGTERM)

	daemon.Daemon()

	time.Sleep(60 * time.Second)
}

func sigtermMyHandler(sig os.Signal) error {
	log.Println("自定义信号量关闭")
	os.Exit(1)
	return daemon.ErrStop
}
