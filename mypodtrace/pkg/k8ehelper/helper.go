package k8ehelper

const (
	NodeUnreachablePodReason = "NodeLost"
	// NodeUnreachablePodMessage is the message on a pod when its state cannot be confirmed as kubelet is unresponsive
	// on the node it is (was) running.
	NodeUnreachablePodMessage = "Node %v which was running pod %v is unresponsive"

	APIGroupAll    = "*"
	ResourceAll    = "*"
	VerbAll        = "*"
	NonResourceAll = "*"

	GroupKind          = "Group"
	ServiceAccountKind = "ServiceAccount"
	UserKind           = "User"

	// AutoUpdateAnnotationKey is the name of an annotation which prevents reconciliation if set to "false"
	AutoUpdateAnnotationKey = "rbac.authorization.kubernetes.io/autoupdate"
)
