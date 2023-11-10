package cmd

import (
	"context"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server(g.Cfg().MustGet(ctx, "server.name").String())
			/*s.BindMiddlewareDefault(middleware.MiddlewareDefaultCORS, middleware.MiddlewareRequestIpLimit)
			xcasbin.CreateCasbinEnforcer(ctx)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.New(),
				)
				group.Group("/api", func(group *ghttp.RouterGroup) {
					group.Bind(user.UserCtrl)
				})
			})
			if err := xgtoken.MyGtoken.New(ctx, gtokenService.GFtokenFnInstance).Start(); err != nil {
				panic(err)
			}*/
			s.Run()
			return nil
		},
	}
)
