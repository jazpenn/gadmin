package router

import (
	"fmt"

	"github.com/gogf/gf/g/net/ghttp"
	"github.com/gogf/gf/g/os/glog"
	"github.com/gogf/gf/g/util/gconv"
	"github.com/hailaz/gadmin/app/api"
	"github.com/hailaz/gadmin/app/model"
	"github.com/hailaz/gadmin/library/common"
)

var routerMap = make(map[string]model.RolePolicy)

func showURL(r *ghttp.Request) {
	glog.Debug(r.Request.RequestURI)
	//r.Response.CORSDefault()
}

func InitRouter(s *ghttp.Server) {

	s.BindHookHandler("/*", ghttp.HOOK_BEFORE_SERVE, showURL)
	s.BindHandler("GET:/loginkey", api.GetLoginCryptoKey)
	s.BindHandler("POST:/login", api.GfJWTMiddleware.LoginHandler)
	s.BindHandler("GET:/refresh_token", api.GfJWTMiddleware.RefreshHandler)
	InitV1(s)

	model.ReSetPolicy("system", routerMap)
}

// authHook is the HOOK function implements JWT logistics.
func authHook(r *ghttp.Request) {
	r.Response.CORSDefault()
	api.GfJWTMiddleware.MiddlewareFunc()(r)
}

// InitV1 初始化V1
//
// createTime:2019年04月25日 09:24:06
// author:hailaz
func InitV1(s *ghttp.Server) {
	//v1 := s.Group("/v1")
	//权限验证
	s.Group("/v1").ALL("/*any", authHook, ghttp.HOOK_BEFORE_SERVE)
	userCtrl := new(api.UserController)
	roleCtrl := new(api.RoleController)
	policyCtrl := new(api.PolicyController)

	// user
	BindGroup(s, "/v1", []ghttp.GroupItem{
		// 用户
		{"POST", "/user/logout", userCtrl, "Logout", "false"},
		{"GET", "/user/info", userCtrl, "Info", "false"},
		{"REST", "/user", userCtrl},
		// 角色
		{"REST", "/role", roleCtrl},
		{"PUT", "/role/byuser", roleCtrl, "SetRoleByUserName"},
		// 权限
		{"REST", "/policy", policyCtrl},
		{"GET", "/policy/byrole", policyCtrl, "GetPolicyByRole"},
		{"PUT", "/policy/byrole", policyCtrl, "SetPolicyByRole"},
	})
}

// BindGroup path string,description
//
// createTime:2019年04月29日 16:45:55
// author:hailaz
func BindGroup(s *ghttp.Server, path string, items []ghttp.GroupItem) {
	g := s.Group(path)
	g.Bind(items)
	for _, item := range items {
		glog.Debug(gconv.String(item[1]))
		if len(item) > 4 && gconv.String(item[4]) == "false" { //不走权限的api
			addPolicy("*", path+gconv.String(item[1]), common.GetAction(gconv.String(item[0])))
		} else { //走权限的api
			if gconv.String(item[0]) == "REST" { //rest api
				addPolicy("system", path+gconv.String(item[1]), model.ACTION_GET)
				addPolicy("system", path+gconv.String(item[1]), model.ACTION_POST)
				addPolicy("system", path+gconv.String(item[1]), model.ACTION_PUT)
				addPolicy("system", path+gconv.String(item[1]), model.ACTION_DELETE)
			} else {
				addPolicy("system", path+gconv.String(item[1]), common.GetAction(gconv.String(item[0])))
			}
		}

	}

}

// addPolicy 记录需要系统路由
//
// createTime:2019年04月29日 17:18:25
// author:hailaz
func addPolicy(role, path, atc string) {
	routerMap[fmt.Sprintf("%v %v %v", role, path, atc)] = model.RolePolicy{Role: role, Path: path, Atc: atc}
}
