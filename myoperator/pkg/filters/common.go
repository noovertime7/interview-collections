package filters

import (
	"github.com/valyala/fasthttp"
	"reflect"
)

type ProxyFilter interface {
	SetValue(value ...string)
	SetPathReg(value ...string)
	Do(ctx *fasthttp.RequestCtx)
}

type ProxyFilters []ProxyFilter //接口、别名 不需要加指针类型

func (this ProxyFilters) Do(ctx *fasthttp.RequestCtx) {
	for _, filter := range this {
		filter.Do(ctx)
	}
}

// 解析配置文件 获取注解值 塞到filter里
func (this ProxyFilters) SetValue(value ...string) {
	for _, filter := range this {
		filter.SetValue(value...)
	}
}

// 将macth到的正则表达式塞到所有的filter中去
func (this ProxyFilters) SetPathReg(pathReg string) {
	for _, filter := range this {
		filter.SetPathReg(pathReg)
	}
}

// 所有的filter需要定义init函数加入到此列表
// 解析配置文件时会根据配置的注解来返回需要使用到的filter
var filterList = map[string]ProxyFilter{}         //作用与请求
var filterListResponse = map[string]ProxyFilter{} //作用于响应

func registerFilter(key string, filter ProxyFilter) {
	filterList[key] = filter
}

func registerFilterResponese(key string, filter ProxyFilter) {
	filterListResponse[key] = filter
}

func CheckAnnotations(anno map[string]string, isRequest bool, ext ...string) (res []ProxyFilter) {
	var list map[string]ProxyFilter
	if isRequest {
		list = filterList
	} else {
		list = filterListResponse
	} //判断是请求过滤器 还是 响应过滤器
	for anno_key, anno_value := range anno {
		for filter_key, filter_value := range list {
			if anno_key == filter_key {
				//tips 通过反射创造一个新的对象 new需要指定类型， 反射不需要
				t := reflect.TypeOf(filter_value)
				if t.Kind() == reflect.Ptr {
					t = t.Elem() //如果是指针就取指向值
				}
				filter := reflect.New(t).Interface().(ProxyFilter)
				params := []string{anno_value}
				params = append(params, ext...)
				filter.SetValue(params...)
				//塞参数
				res = append(res, filter)
			}
		}
	}
	return
}
