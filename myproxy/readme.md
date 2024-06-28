a multi cluster k8s management tool base on official k8s client-go & kubectl proxy

details
1.apply请求会先get一遍，客户端根据结果来决定是patch还是post
2.get资源会先请求一遍api-resource，确认请求的资源类型是否存在
3.kubectl get xxx --selector="xxx"实际上会在请求的query参数中加上labelSelector=xxx
