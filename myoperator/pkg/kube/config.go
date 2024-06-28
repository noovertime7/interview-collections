package kube

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func GetConfig() *rest.Config {
	cfg, err := clientcmd.BuildConfigFromFlags("", "./resources/config")
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
