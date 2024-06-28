package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/mycore"
	"path/filepath"
)

func GetCliSet() (*kubernetes.Clientset, error) {
	file := "e-config"
	configPath := filepath.Join(clientcmd.RecommendedConfigDir, file)
	cfg, err := clientcmd.BuildConfigFromFlags("", configPath) //可以用kubelet.config
	if err != nil {
		return nil, err
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func main() {
	cli, _ := GetCliSet()

	mycore.RegisterNode(cli, "mynode")
}
