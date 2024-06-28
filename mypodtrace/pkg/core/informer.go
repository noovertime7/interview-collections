package core

import (
	"fmt"
	"github.com/octoboy233/mypodtrace/pkg/config"
	"github.com/octoboy233/mypodtrace/pkg/handlers"
	"k8s.io/client-go/informers"
	"os"
	"os/signal"
)

//var PodLister v1.PodLister

// 启动informer监听pod
func WatchPod() {
	client := config.InitClientSet()
	fact := informers.NewSharedInformerFactory(client, 0)

	// pod informer
	podInformer := fact.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(handlers.NewPodHandlers())
	// event informer
	eventInformer := fact.Core().V1().Events().Informer()
	eventInformer.AddEventHandler(handlers.NewEventHandlers())
	// deployment informer
	deploymentInformer := fact.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(handlers.NewDeploymentHandlers())
	// rs informer
	rsInformer := fact.Apps().V1().ReplicaSets().Informer()
	rsInformer.AddEventHandler(handlers.NewReplicaSetHandlers())

	ch := make(chan struct{})
	fmt.Println("k8s可观测启动...")
	fact.Start(ch)

	//PodLister = fact.Core().V1().Pods().Lister()

	//hang
	notifyCh := make(chan os.Signal)
	signal.Notify(notifyCh, os.Interrupt, os.Kill)
	<-notifyCh
	fmt.Println("k8s可观测停止...")
}
