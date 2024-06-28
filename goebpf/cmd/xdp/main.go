package main

import "goebpf/cebpf/xdp"

func main() {
	xdp.LoaderTcWrite()
}
