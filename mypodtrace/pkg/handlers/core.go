package handlers

import (
	"context"
	lru "github.com/hashicorp/golang-lru/v2"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

type spanInfo struct {
	rootCtx context.Context //仅为pod做区分使用
	ctx     context.Context //pod更新作为ctx的二级span，event生成作为rootCtx的二级span（与ctx同级）
}

var (
	CtxMapper *lru.Cache[types.UID, spanInfo]
)

func init() {
	set, err := lru.New[types.UID, spanInfo](1280)
	if err != nil {
		panic(err)
	}
	CtxMapper = set
}

func IsTestResource(name string) bool {
	if strings.Contains(name, "test") {
		return true
	}
	return false
}
