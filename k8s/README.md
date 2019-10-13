# K8S部署

## plugin.go增加`kubernetes`注册中心插件
```go
package main

import (
	// k8s registry
	_ "github.com/micro/go-plugins/registry/kubernetes"
)
```