package namespace

import (
	v1 "k8s.io/api/core/v1"
	"sort"
	"sync"
)

type NsMap struct {
	data sync.Map //key cluster value []*corev1.namespace
}

func (this *NsMap) Add(cluster string, ns *v1.Namespace) {
	if value, ok := this.data.Load(cluster); ok {
		if nsList, ok := value.([]*v1.Namespace); ok {
			nsList = append(nsList, ns)
			this.data.Store(cluster, nsList)
		}
	} else {
		this.data.Store(cluster, []*v1.Namespace{ns})
	}
}

func (this *NsMap) Update(cluster string, ns *v1.Namespace) {
	if value, ok := this.data.Load(cluster); ok {
		if nsList, ok := value.([]*v1.Namespace); ok {
			for i, ns_in_list := range nsList {
				if ns_in_list.UID == ns.UID {
					nsList[i] = ns
					this.data.Store(cluster, nsList)
				}
			}
		}
	}
}

func (this *NsMap) Delete(cluster string, ns *v1.Namespace) {
	if value, ok := this.data.Load(cluster); ok {
		if nsList, ok := value.([]*v1.Namespace); ok {
			for i, ns_in_list := range nsList {
				if ns_in_list.UID == ns.UID && ns_in_list.Name == ns.Name {
					nsList = append(nsList[:i], nsList[i+1:]...)
					this.data.Store(cluster, nsList)
				}
			}
		}
	}
}

func (this *NsMap) List(cluster string) []*Ns {
	if value, ok := this.data.Load(cluster); ok {
		list := NsItems(value.([]*v1.Namespace))
		sort.Sort(list)
		res := make([]*Ns, len(list))
		for i, ns := range list {
			res[i] = &Ns{Name: ns.Name}
		}
		return res
	}
	return []*Ns{}
}

type NsItems []*v1.Namespace

func (this NsItems) Len() int {
	return len(this)
}

func (this NsItems) Less(i, j int) bool {
	return this[i].Name < this[j].Name
}

func (this NsItems) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
