package node

import (
	"context"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/component-helpers/apimachinery/lease"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"os"
	"time"
)

//node租约 保持节点在线

const (
	leaseDuration = 40
	renewInterval = time.Duration(leaseDuration * 0.25)
)

func onRepeatedHeartbeatFailure() {
	//就是一些更新租期失败时的清理操作
	klog.Fatalln("onRepeatedHeartbeatFailure")
	os.Exit(1)
}

// SetNodeOwnerFunc helps construct a newLeasePostProcessFunc which sets
// a node OwnerReference to the given lease object
func SetNodeOwnerFunc(c clientset.Interface, nodeName string) func(lease *coordinationv1.Lease) error {
	return func(lease *coordinationv1.Lease) error {
		// Setting owner reference needs node's UID. Note that it is different from
		// kubelet.nodeRef.UID. When lease is initially created, it is possible that
		// the connection between master and node is not ready yet. So try to set
		// owner reference every time when renewing the lease, until successful.
		if len(lease.OwnerReferences) == 0 {
			if node, err := c.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{}); err == nil {
				lease.OwnerReferences = []metav1.OwnerReference{
					{
						APIVersion: corev1.SchemeGroupVersion.WithKind("Node").Version,
						Kind:       corev1.SchemeGroupVersion.WithKind("Node").Kind,
						Name:       nodeName,
						UID:        node.UID,
					},
				}
			} else {
				klog.ErrorS(err, "Failed to get node when trying to set owner ref to the node lease", "node", klog.KRef("", nodeName))
				return err
			}
		}
		return nil
	}
}

// starts a lease controller for the node.
// create lease if not exist and keep updating lease
func StartLeaseController(cli clientset.Interface, nodeName string) {
	timer := clock.RealClock{} // 封装了time.Now()等方法的接口
	ctl := lease.NewController(timer,
		cli,
		nodeName,
		leaseDuration,
		onRepeatedHeartbeatFailure,
		renewInterval,
		nodeName,
		corev1.NamespaceNodeLease,
		SetNodeOwnerFunc(cli, nodeName),
	)
	klog.Infoln("start lease controller")
	ctl.Run(context.Background())
}

// 更新节点租期
// low b 方法，for循环然后手工更新
func RenewNode(cli clientset.Interface, lease *coordinationv1.Lease) error {
	now := metav1.NewMicroTime(time.Now())
	lease.Spec.RenewTime = &now
	newLease, err := cli.CoordinationV1().Leases(corev1.NamespaceNodeLease).Update(context.Background(), lease, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	lease = newLease
	return nil
}

// 将mynode置为ready状态, 该步需要配合 节点续期 一起操作，否则会被controller-manager立刻检测到 并标记为notReady
func UpdateNodeStatus(cli clientset.Interface, nodeName string) error {
	node, _ := cli.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	status := node.Status //tips: 1.status成员变量不是指针类型，需要重新赋值回去
	for i, cond := range status.Conditions {
		if cond.Type == corev1.NodeReady {
			cond.Status = corev1.ConditionTrue
			cond.Reason = "KubeletReady"
			cond.Message = "kubelet is ready"
			status.Conditions[i] = cond // tips：同理，status.Conditions[i]不是指针类型，需要重新赋值回去·
		}
	}
	node.Status = status // 1.
	_, err := cli.CoreV1().Nodes().UpdateStatus(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
