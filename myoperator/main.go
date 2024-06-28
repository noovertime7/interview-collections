package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/yeqown/log"
	"jtproxy/pkg/sysinit"
)

func main() {
	sysinit.InitConfig()
	log.Fatal(fasthttp.ListenAndServe(fmt.Sprintf(":%d", sysinit.SysConfig.Server.Port), ProxyHandler))
}
