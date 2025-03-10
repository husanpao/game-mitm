package main

import (
	"fmt"
	gamemitm "github.com/husanpao/game-mitm"
	"github.com/husanpao/game-mitm/gosysproxy"

	"os"
	"os/signal"
	"syscall"
)

func init() {
	err := gosysproxy.SetGlobalProxy(
		"127.0.0.1:12311",
		"localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;192.168.*",
	)
	if err != nil {
		panic(err)
	}
}
func main() {
	proxy := gamemitm.NewProxy()
	proxy.SetVerbose(true)

	proxy.OnRequest("echo.websocket.events").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		fmt.Println("OnRequest")
		return body
	})

	proxy.OnResponse("echo.websocket.events").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		fmt.Println("OnResponse")
		return body
	})
	proxy.OnConnected("echo.websocket.events").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		fmt.Println("OnConnected")
		ctx.WSSession.SendTextToServer([]byte("777777777"))
		return body
	})

	// 监听操作系统信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动代理
	go proxy.Start()

	// 等待程序终止信号
	<-signalChan
	gosysproxy.Off()
	// 在程序结束时执行清理操作

}
