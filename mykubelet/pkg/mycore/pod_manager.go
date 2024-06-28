package mycore

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	corev1 "k8s.io/kubernetes/pkg/apis/core/v1"
	"k8s.io/kubernetes/pkg/kubelet/config"
	"k8s.io/kubernetes/pkg/kubelet/configmap"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
	"k8s.io/kubernetes/pkg/kubelet/secret"
	"k8s.io/kubernetes/pkg/kubelet/status"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/utils/clock"
	"reflect"
)

// 就是官方的 PodManager  以它为主体再带上一些必要成员变量， 构成精简版的kubelet
type PodCache struct {
	cli        clientset.Interface
	PodManager kubepod.Manager   //主要用来存pod对象 内置informerFactory
	PodConfig  *config.PodConfig //主要和syncLoop交互
	Clock      clock.RealClock
	PodWorker  PodWorkers          //用于开启协程来调用cri和同步podStatus等操作
	InnerCache kubecontainer.Cache //用于构造worker的cache。发生变化会触发syncPod
}

func NewPodCache(cli clientset.Interface, nodeName string) *PodCache {
	var pm kubepod.Manager

	eventBroadcaster := record.NewBroadcaster()

	//事件广播器指定发送目的端
	_ = corev1.AddToScheme(legacyscheme.Scheme)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: cli.CoreV1().Events("")})

	recorder := eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: "kubelet", Host: nodeName})

	fac := informers.NewSharedInformerFactory(cli, 0)
	fac.Core().V1().Nodes().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
	ch := make(chan struct{})
	fac.Start(ch)
	if waitMap := fac.WaitForCacheSync(ch); waitMap[reflect.TypeOf(&v1.Node{})] {
		nodeLister := fac.Core().V1().Nodes().Lister()
		mirrorPodClient := kubepod.NewBasicMirrorClient(cli, "mymbp", nodeLister)
		secretManager := secret.NewSimpleSecretManager(cli)
		configMapManager := configmap.NewSimpleConfigMapManager(cli)

		pm = kubepod.NewBasicPodManager(mirrorPodClient, secretManager, configMapManager)
	}

	cl := clock.RealClock{}
	innerCache := kubecontainer.NewCache()

	statusManager := status.NewManager(cli, pm, &PodDeletionSafetyProviderStruct{})
	statusManager.Start()

	podFn := NewPodFn(cli, statusManager, recorder)

	return &PodCache{
		cli:        cli,
		PodManager: pm,
		PodConfig:  NewPodConfig(cli, nodeName, fac, recorder),
		Clock:      cl,
		PodWorker:  NewPodWorkers(recorder, cl, innerCache, podFn, pm),
		InnerCache: innerCache,
	}
}

func NewPodConfig(cli clientset.Interface, nodeName string,
	fac informers.SharedInformerFactory, recorder record.EventRecorder) *config.PodConfig {

	// source of all configuration
	podCfg := config.NewPodConfig(config.PodConfigNotificationIncremental, recorder)
	//Adding apiserver pod source
	config.NewSourceApiserver(cli, types.NodeName(nodeName), func() bool {
		return fac.Core().V1().Nodes().Informer().HasSynced()
	}, podCfg.Channel(kubetypes.ApiserverSource))

	return podCfg
}

// PodDeletionSafetyProviderStruct 搞个假的
type PodDeletionSafetyProviderStruct struct{}

func (p PodDeletionSafetyProviderStruct) PodResourcesAreReclaimed(pod *v1.Pod, status v1.PodStatus) bool {
	return true

}

// 不断通过容器运行时检查容器是否被删除了
func (p PodDeletionSafetyProviderStruct) PodCouldHaveRunningContainers(pod *v1.Pod) bool {
	return true
}

var _ status.PodDeletionSafetyProvider = &PodDeletionSafetyProviderStruct{}
