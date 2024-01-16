package http

import (
	"github.com/go-kratos/kratos/v2/transport/http"
)

// RouteAppender krstos 路由追加器
type RouteAppender interface {
	// 追加路由到kratos HTTP Server
	AppendToServer(srv *http.Server)
}
