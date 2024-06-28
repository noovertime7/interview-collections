package resources

type Resources struct {
	Name  string
	Verbs []string
}

// 分类 同一个gv下所有的resource
type GroupResources struct {
	Group     string
	Version   string
	Resources []*Resources
}

type clusterList []string

func (c clusterList) Len() int {
	return len(c)
}

func (c clusterList) Less(i, j int) bool {
	return c[i] < c[j]
}

func (c clusterList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
