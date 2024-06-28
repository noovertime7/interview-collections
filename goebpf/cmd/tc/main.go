package main

import (
	"fmt"
	"goebpf/cebpf/tc"
)

func main() {
	fmt.Println("启动ebpf")
	tc.LoaderTcWrite()
}
