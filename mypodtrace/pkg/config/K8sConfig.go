package config

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

const CONFIG_PATH = "~/.kube/config"

// 初始化client-go的clientSet
func InitClientSet() *kubernetes.Clientset {
	cfg, err := clientcmd.BuildConfigFromFlags("", CONFIG_PATH)
	if err != nil {
		panic(err)
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	return cli
}

// 获取kubeconfig里的集群名
func GetClusterName() string {
	configFlags := genericclioptions.NewConfigFlags(true)
	cliConf := configFlags.ToRawKubeConfigLoader()

	apiConfig, err := cliConf.RawConfig()
	if err != nil {
		log.Fatal(err)
	}
	return apiConfig.Contexts[apiConfig.CurrentContext].Cluster
}

// 获取集群版本信息
func GetClusterVersion() string {
	cli := InitClientSet()
	versionInfo, err := cli.ServerVersion()
	if err != nil {
		log.Fatal(err)
	}
	return versionInfo.GitVersion
}
