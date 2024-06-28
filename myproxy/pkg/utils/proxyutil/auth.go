package proxyutil

import (
	"encoding/base64"
	"fmt"
	"github.com/casbin/casbin/v2"
	"k8s.io/klog/v2"
	"myproxy/pkg/dbcore"
	"net/http"
	"regexp"
	"strings"
)

const (
	model_path  = "./resources/casbin/model.conf"
	policy_path = "./resources/casbin/policy.csv"
)

var enforcer *casbin.Enforcer

func init() {
	// 因为policy做持久化需要 配合后台系统调用 casbin API去做增删改茶，所以 暂时用csv文件代替
	enforcer, _ = casbin.NewEnforcer(model_path, policy_path)
}

func auth(r *http.Request, user string) bool {
	if regexp.MustCompile("^/apis/apps/v1/namespaces/[^/]+/deployments(?:/[^?]+)?(?:\\?.*)*$").
		MatchString(r.RequestURI) { // 暂时 只验证deploy
		pass, err := enforcer.Enforce(user, r.URL.Path, r.Method)
		if err != nil {
			fmt.Println(err)
		}
		return pass
	}
	return true
}

// 从请求中获取认证信息
func extractToken(req *http.Request) (info []string) {
	if auths, ok := req.Header["Authorization"]; ok {
		for _, auth := range auths {
			pair := strings.Split(auth, " ")
			if len(pair) == 2 && (pair[0] == "Bearer" || pair[0] == "Basic") {
				token := pair[1]
				s, _ := base64.StdEncoding.DecodeString(token)
				info = append(info, strings.Split(string(s), ":")...)
			}
		}
	}
	return info
}

func getUser(username string) *dbcore.User {
	user := &dbcore.User{}
	dbcore.DB.Table("user").Where("user = ?", username).Find(user)
	return user
}

// Auth 传入 http Header 返回鉴权结果
// 结合casbin做权限控制
// todo 在表中创建盐值字段，对密码进行加盐加密，将用户传入的密码进行加盐加密后与数据库中的密码散列值进行比对，避免在数据库中存储明文密码
func Auth(req *http.Request) bool {
	//从请求中获取认证信息
	info := extractToken(req)
	// expect info  []string{user1,123456}
	if len(info) != 2 {
		// 未包含认证信息
		klog.Info("接受到了未包含授权信息的请求: " + req.Method + "  " + req.RequestURI)
		return false
	} else {
		u := getUser(info[0])

		if u.Passwd == info[1] { // 通过密码校验
			if ok := auth(req, u.User); ok {
				klog.Info("接受到了来自 " + u.User + " 的请求: " + req.Method + "  " + req.RequestURI)
				return ok
			} else {
				klog.Info("接受到了来自 " + u.User + " 的请求, 鉴权未通过: " + req.Method + "  " + req.RequestURI)
				return ok
			}
		} else { // dismatch
			klog.Info("接受到了来自 " + info[0] + " 的请求, 密码不正确: " + req.Method + "  " + req.RequestURI)
			return false
		}
	}
}
