package main

import (
	"crypto/tls"
	"log"
	"myproxy/pkg/utils/proxyutil"
	"net"
)

func HttpListener(server *proxyutil.Server) net.Listener {
	//http://localhost:8009/api/v1/namespaces/default/pods
	listen, err := server.Listen("0.0.0.0", 8009)
	if err != nil {
		log.Fatal(err)
	}
	return listen
}

func HttpsListener() net.Listener {
	//keypoint 为什么要用https的方式启动代理呢？
	//keypoint 因为kubeconfig中如果指定的server地址是http地址的话，用kubectl配置这个config访问，不会将barrier-token放进header
	//keypoint 这里的需求场景是这样：分发不同的kubeconfig，里面的token各不相同，代理拦截请求获得header，根据token，结合casbin做权限控制的设计
	//keypoint 如果是多集群的需求，简单的用http也可以
	cert, err := tls.LoadX509KeyPair("./certs/server.pem",
		"./certs/server-key.pem")
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	addr := "0.0.0.0:8009"
	listener, err := tls.Listen("tcp", addr, &tlsConfig)
	if err != nil {
		log.Fatal(listener)
	}
	return listener
}

// proxy二开
func main() {
	// "/" + "/static" 访问 "./test/proxy"中的静态文件
	//server, err := proxyutil.NewServer("./test/proxy", "/", "/static",
	//	nil, config.NewK8sConfig().K8sRestConfigDefault(), 0, false)
	server, err := proxyutil.NewServerUponMultiCluster("./test/proxy", "/", "/static",
		nil, 0, false)
	if err != nil {
		log.Fatal(err)
	}
	// 寻找http包中的handler和transport去包装来做请求的拦截
	// handler只能拦截到请求，transport可以做更深层次的拦截，还可以获取到response，并对其内容进行加工
	//server.ServeOnListener(HttpListener(server))
	err = server.ServeOnListener(HttpsListener())
	if err != nil {
		log.Fatal(err)
	}

}
