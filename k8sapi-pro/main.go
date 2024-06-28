package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	"k8sapi-pro/pkg/configmap"
	"k8sapi-pro/pkg/core"
	"k8sapi-pro/pkg/deployment"
	"k8sapi-pro/pkg/ingress"
	"k8sapi-pro/pkg/namespace"
	"k8sapi-pro/pkg/node"
	"k8sapi-pro/pkg/pod"
	"k8sapi-pro/pkg/rbac"
	"k8sapi-pro/pkg/resources"
	"k8sapi-pro/pkg/secret"
	"k8sapi-pro/pkg/service"
	"k8sapi-pro/pkg/user"
	"k8sapi-pro/pkg/websocket"
	"k8sapi-pro/src/config"
	"net/http"
)

func cros() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization,X-Token")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

	}
}

func main() {
	goft.Ignite(cros()).
		Config(
			config.NewDbConfig(),
			config.NewK8sHandler(),
			config.NewK8sConfig(),
			config.NewK8sService(),
			//config.NewK8sMap(),
		).
		Mount("",
			websocket.NewWsCtl(),
			user.NewUserCtl(),
			resources.NewResourcesCtl(),
			core.NewCoreCtl(),
		).
		Mount("/v1",
			deployment.NewDeploymentCtl(),
			pod.NewPodCtl(),
			namespace.NewNsCtl(),
			ingress.NewIngressCtl(),
			service.NewServiceCtl(),
			configmap.NewConfigMapCtl(),
			secret.NewSecretCtl(),
			node.NewNodeCtl(),
			rbac.NewRBACCtl(),
		).
		Launch()
}
