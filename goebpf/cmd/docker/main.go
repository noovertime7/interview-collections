package main

import "goebpf/cebpf/docker"

func main() {
	//docker.LoaderVethXDP() //加载veth的xdp
	//docker.ClearTC()  // 清除tc
	docker.LoaderTC() // 加载tc
}
