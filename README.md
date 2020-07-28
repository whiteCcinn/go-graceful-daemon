# go-graceful-daemon

Make it easier for your programs to become daemons

## usage

```go
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

```

All you need to do is introduce `daemon "github.com/whiteCcinn/go-graceful-daemon"` and `daemon.Daemon()`

