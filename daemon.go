package go_graceful_daemon

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var (
	AppName string
	PidFile string
	PidVal  int
)

func Daemon() {
	file, _ := filepath.Abs(os.Args[0])
	appPath := filepath.Dir(file)
	AppName = filepath.Base(file)
	pidFile, exist := os.LookupEnv("GO_GRACEFUL_DAEMON_PID_PATH")
	if !exist {
		pidFile = appPath + "/" + AppName + ".pid"
	} else {
		pidFile += ".pid"
	}
	PidFile = pidFile
	if os.Getenv("__Daemon") != "true" { //master
		cmd := "start" //缺省为start
		if l := len(os.Args); l > 1 {
			cmd = os.Args[l-1]
		}
		switch cmd {
		case "start":
			if isRunning() {
				log.Printf("[%d] %s is running\n", PidVal, AppName)
			} else { // fork daemon进程
				if err := forkDaemon(); err != nil {
					log.Fatal(err)
				}
			}
		case "restart": // 重启:
			if !isRunning() {
				log.Printf("%s not running\n", AppName)
			} else {
				log.Printf("[%d] %s restart now\n", PidVal, AppName)
				restart(PidVal)
			}
		case "stop": // 停止
			if !isRunning() {
				log.Printf("%s not running\n", AppName)
			} else {
				syscall.Kill(PidVal, syscall.SIGTERM) //kill
			}
		case "-h":
			fmt.Printf("Usage: %s start|restart|stop\n", AppName)
		default: //其它不识别的参数
			return //返回至调用方
		}
		//主进程退出
		os.Exit(0)
	}
	go handleSignals()
}

//检查PidFile是否存在以及文件里的pid是否存活
func isRunning() bool {
	if mf, err := os.Open(PidFile); err == nil {
		pid, _ := ioutil.ReadAll(mf)
		PidVal, _ = strconv.Atoi(string(pid))
	}
	running := false
	if PidVal > 0 {
		if err := syscall.Kill(PidVal, 0); err == nil { // 发一个信号为0到指定进程ID，如果没有错误发生，表示进程存活
			running = true
		}
	}
	return running
}

//保存pid
func savePid(pid int) error {
	file, err := os.OpenFile(PidFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(strconv.Itoa(pid))
	return nil
}

// 捕获系统信号
func handleSignals() {
	err := ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}

// forkDaemon,当checkPid为true时，检查是否有存活的，有则不执行
func forkDaemon() error {
	args := os.Args
	os.Setenv("__Daemon", "true")
	procAttr := &syscall.ProcAttr{
		Env:   os.Environ(),
// 		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	pid, err := syscall.ForkExec(os.Args[0], args, procAttr)
	if err != nil {
		return err
	}
	log.Printf("[%d] %s start daemon\n", pid, AppName)
	savePid(pid)
	return nil
}

//重启(先发送kill -HUP到运行进程，手工重启daemon ...当有运行的进程时，daemon不启动)
func restart(pid int) {
	syscall.Kill(pid, syscall.SIGHUP) //kill -HUP, daemon only时，会直接退出
	fork := make(chan bool, 1)
	go func() { // 循环，查看PidFile是否存在，不存在或值已改变，发送消息
		for {
			f, err := os.Open(PidFile)
			if err != nil || os.IsNotExist(err) { //文件已不存在
				fork <- true
				break
			} else {
				PidVal, _ := ioutil.ReadAll(f)
				if strconv.Itoa(pid) != string(PidVal) {
					fork <- false
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	//处理结果
	select {
	case r := <-fork:
		if r {
			forkDaemon()
		}
	case <-time.After(time.Second * 5):
		log.Fatalln("restart timeout")
	}
}
