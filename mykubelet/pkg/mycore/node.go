package mycore

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"runtime"
)

func RegisterNode(cli clientset.Interface, nodeName string) {
	// check if node exist
	_, err := cli.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})

	if err != nil { // node not exist
		// mock node info
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
				Labels: map[string]string{
					"kubernetes.io/hostname": nodeName,
					"kubernetes.io/os":       runtime.GOOS,
					"kubernetes.io/arch":     runtime.GOARCH,
				},
			},
			Spec: v1.NodeSpec{},
		}
		_, err = cli.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
		if err != nil {
			klog.Errorln(err)
		}
	} else {
		klog.Infof("node %s already exist", nodeName)
	}

	//err = UpdateNodeStatus(cli, nodeName) //set node ready temporarily
	//if err != nil {
	//	klog.Errorln(err)
	//}
	return
}
