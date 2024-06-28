package models

import (
	"crypto/md5"
	"fmt"
	"github.com/shenyisyn/goft-gin/goft"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"
	"time"
)

type Resources struct {
	obj             *unstructured.Unstructured `gorm:"-"`
	objbytes        []byte                     `gorm:"-"`
	Id              int                        `gorm:"column:id;primaryKey;autoIncrement"`
	NameSpace       string                     `gorm:"column:namespace"`
	Name            string                     `gorm:"column:name"`
	ResourceVersion string                     `gorm:"column:resource_version"`
	Hash            string                     `gorm:"column:hash"`
	//-----------gvr 和kind相关
	Group    string `gorm:"column:group"`
	Version  string `gorm:"column:version"`
	Resource string `gorm:"column:resource"`
	Kind     string `gorm:"column:kind"`
	//-------end grv
	// TODO
	//-- uid是唯一的 。owner 不一定有
	Owner string `gorm:"column:owner"`
	Uid   string `gorm:"column:uid"`
	//-- end uid and owner
	Object string `gorm:"column:object"`

	// ----时间相关
	CreateAt time.Time `gorm:"column:create_at"`

	UpdateAt time.Time `gorm:"column:update_at"`

	DeleteAt time.Time `gorm:"column:delete_at"`
	//--UpdateAt DeleteAt 插入时 不需要赋值

	Cluster string `gorm:"column:cluster"`
}

func NewResource(cluster string, obj runtime.Object, mapper meta.RESTMapper) (*Resources, error) {
	o := &unstructured.Unstructured{}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, o)
	if err != nil {
		return nil, err
	}

	r := &Resources{Cluster: cluster, obj: o, objbytes: b}
	gvk := o.GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version) //tips 获取gvr的方法
	if err != nil {
		return nil, err
	}
	r.Group = gvk.Group
	r.Version = gvk.Version
	r.Kind = gvk.Kind
	r.Resource = mapping.Resource.Resource

	r.NameSpace = o.GetNamespace()
	r.Name = o.GetName()
	r.ResourceVersion = o.GetResourceVersion()
	r.Uid = string(o.GetUID())
	r.CreateAt = o.GetCreationTimestamp().Time

	// 一般mysql的设置不允许插入零值时间，需要修改数据库的配置文件。因此 选取一个时间作为默认时间
	r.UpdateAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	r.DeleteAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	return r, nil
}

func (this *Resources) prepare() {
	//设置owner
	owner := this.obj.GetOwnerReferences()
	if len(owner) > 0 {
		this.Owner = string(owner[0].UID)
	}
	//设置hash值，防止重启时重复插入
	this.Hash = fmt.Sprintf("%x", md5.Sum(this.objbytes))
	obj_json, err := yaml.YAMLToJSON(this.objbytes)
	goft.Error(err)
	this.Object = string(obj_json)

	goft.Error(this.transEvent())
}

// 针对event的特殊处理
// 1.修改了记录的name，方便查询pod信息时根据pod名称获取事件
// 2.一个pod对应多个事件，必然存在重复的name，为了方便查询，覆盖了事件最近发生时间，查询时降序取第一条事件记录
func (this *Resources) transEvent() error {
	if this.Kind == "Event" {
		invol := this.obj.Object["involvedObject"].(map[string]interface{})
		involName := invol["name"].(string)
		involKind := invol["kind"].(string)
		this.Name = fmt.Sprintf("%s_%s", involKind, involName)
		last := this.obj.Object["lastTimestamp"]
		if last != nil {
			to, err := time.Parse("2006-01-02T15:04:05Z", last.(string))
			if err != nil {
				return err
			}
			this.UpdateAt = to
		}
	}
	return nil
}

// tips 程序重启会把所有资源触发一遍add，所以要去重
func (this *Resources) Add(db *gorm.DB) error {
	this.prepare()
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uid"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"update_at": time.Now(),
		}),
	}).Create(this).Error
}

// tips configmap会一直触发update 所以要过滤这部分的操作
func (this *Resources) Update(db *gorm.DB) error {
	this.prepare()
	return db.Where("uid = ?", this.Uid).
		Where("hash != ?", this.Hash).
		Updates(this).Error
}

// todo 不直接删除而是更新删除时间戳，做操作审计
func Delete(uid string, db *gorm.DB) error {
	return db.Where("uid = ?", uid).
		Delete(&Resources{}).Error
}

func List(cluster string, ns string, kind string, db *gorm.DB) ([]Resources, error) {
	resources := []Resources{}
	err := db.Where("cluster = ?", cluster).
		Where("namespace = ?", ns).
		Where("kind= ? ", kind).Find(&resources).Error
	return resources, err
}

func ListNoNamespaced(cluster string, kind string, db *gorm.DB) ([]Resources, error) {
	resources := []Resources{}
	err := db.Where("cluster = ?", cluster).
		Where("kind= ? ", kind).Find(&resources).Error
	return resources, err
}

func Take(cluster string, ns string, name string, kind string, db *gorm.DB) (Resources, error) {
	resource := Resources{}
	err := db.Where("cluster = ?", cluster).
		Where("namespace = ?", ns).
		Where("name = ?", name).
		Where("kind= ? ", kind).
		Order("update_at desc"). //查询事件时取最新展示
		Find(&resource).Error
	//Take(&resource).Error
	return resource, err
}

func TakeNoNamespaced(cluster string, name string, kind string, db *gorm.DB) (Resources, error) {
	resource := Resources{}
	err := db.Where("cluster = ?", cluster).
		Where("name = ?", name).
		Where("kind= ? ", kind).
		Find(&resource).Error
	//Take(&resource).Error
	return resource, err
}
