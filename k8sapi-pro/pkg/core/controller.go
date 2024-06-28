package core

import (
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	"k8sapi-pro/src/models"
)

type CoreCtl struct {
	Factories *models.InformerList `inject:"-"`
	GVRs      *models.GVRs         `inject:"-"`
}

func (this *CoreCtl) Name() string {
	return "CoreCtl"
}

func NewCoreCtl() *CoreCtl {
	return &CoreCtl{}
}

func (this *CoreCtl) Build(goft *goft.Goft) {
	goft.Handle("POST", "/wipeout", this.ClearCache)
}

// 清除informer缓存
func (this *CoreCtl) ClearCache(c *gin.Context) goft.Json {
	for _, fac := range *this.Factories {
		for _, gvr := range *this.GVRs {
			store := fac.ForResource(gvr).Informer().GetStore()
			// 遍历对象进行删除
			for _, obj := range store.List() {
				// delete 对象
				if err := store.Delete(obj); err != nil {
					goft.Error(err)
				}
			}
		}
	}

	return gin.H{
		"code":    20000,
		"message": "OK",
	}
}
