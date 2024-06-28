package proxyutil

import (
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"log"
	"os"
)

const (
	CONFIG_PATH = "./resources/kubeconfig/config"
)

func K8sApiConfig() *api.Config {
	configFile, err := os.Open(CONFIG_PATH)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatal(err)
	}
	cc, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		log.Fatal(err)
	}
	ac, err := cc.RawConfig()
	if err != nil {
		log.Fatal(err)
	}
	return &ac
}

type ClusterService struct {
	*api.Config
}

func NewClusterService() *ClusterService {
	return &ClusterService{Config: K8sApiConfig()}
}

type RestConfig struct {
	restConfig *rest.Config
	isDefault  bool
}

func (this *ClusterService) GenerateRestMap() map[string]*RestConfig { //key: context name,value:rest.config
	set := make(map[string]*RestConfig)
	for context_name, _ := range this.Config.Contexts {
		restconf, err := this.GetRestByCtxName(context_name)
		if err != nil {
			err = fmt.Errorf("获取rest.config出错:%s", err)
			fmt.Println(err)
		}
		fmt.Println("加载集群配置，cluster名称：" + context_name)
		set[context_name] = &RestConfig{restConfig: restconf}

		if context_name == this.Config.CurrentContext {
			set[context_name].isDefault = true
		}
	}
	return set
}

func (this *ClusterService) GetRestByCtxName(ctxName string) (*rest.Config, error) {
	restconfig, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
		apiconfig_copy := this.Config.DeepCopy()
		apiconfig_copy.CurrentContext = ctxName //keypoint
		return apiconfig_copy, nil
	})
	if err != nil {
		return nil, err
	}
	return restconfig, err
}
