package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"xhhrobot/ai"
	"xhhrobot/config"
	"xhhrobot/db"
	"xhhrobot/loger"
	"xhhrobot/updater"
	"xhhrobot/xhh"
)

func main() {
	loger.InitLog()
	mode := flag.String("mode", "default", "Switch a mode when start")
	flag.Parse()

	if *mode == "update" {
		if err := updater.Run(); err != nil {
			loger.Loger.Error("[UPDATE]" + err.Error())
		}
		return
	}

	config.InitConfig()
	time.Sleep(1 * time.Second)
	db.Init()
	ai.Init()
	start(mode)
	loger.Loger.Info("[MAIN]正在关闭...")
	ai.Close()
}

func CheckNew() {
	if !db.IsNew() {
		return
	}
	fmt.Println("检测到您是第一次运行\n是否允许将先前@过的名单加入至艾特列表？\ny(es) or n(o) 默认n\n请输入y或n")
	input := bufio.NewReader(os.Stdin)
	str, err := input.ReadString('\n')
	if err != nil {
		loger.Loger.Fatal("[MAIN]无法读取您的输入")
	}
	in := strings.TrimRight(str, "\r\n")

	switch in {
	case "y":
		xhh.DontReply = false
		return
	case "yes":
		xhh.DontReply = false
		return
	case "Y":
		xhh.DontReply = false
		return
	case "Yes":
		xhh.DontReply = false
		return
	case "YES":
		xhh.DontReply = false
		return
	default:
		xhh.DontReply = true
		return
	}
}

func start(mode *string) {
	switch *mode {
	case "default":
		loger.Loger.Info("\nhttps://github.com/ZzzHe2333/xhhRobot\n你需要输入启动项\n-mode start | login | test | update\nWindows 也可以双击 update.bat 一键更新")
	case "test":
		xhh.RunTest()
	case "login":
		xhh.Login()
	case "start":
		CheckNew()
		xhh.Init()
		xhh.Start()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
	}
}
