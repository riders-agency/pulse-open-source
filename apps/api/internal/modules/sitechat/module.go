package sitechat

import (
	"github.com/go-chi/chi/v5"
	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"

	sitechathttp "riders.agency/pulse/api/internal/modules/sitechat/transport/http"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

type Hub = sitechathttp.Hub

func NewHub(siteChat *sitechatsvc.Commands) *Hub {
	return sitechathttp.NewHub(siteChat)
}

func RegisterHTTP(r chi.Router, deps httpserver.Dependencies) {
	MountWidgetHTTP(r)
}

func RegisterAPIV1(scope httpserver.APIV1Scope, deps httpserver.Dependencies) {
	switch scope.Kind {
	case httpserver.APIV1Public:
		MountPublicHTTP(scope.Router, deps.SiteChat, deps.SiteChatHub)
	case httpserver.APIV1Protected:
		MountOperatorHTTP(scope.Router, deps.SiteChat, deps.SiteChatHub)
	case httpserver.APIV1Project:
		MountProjectHTTP(scope.Router, deps.SiteChat)
	}
}

func MountWidgetHTTP(r chi.Router) {
	sitechathttp.MountWidgetScriptRoute(r)
}

func MountPublicHTTP(r chi.Router, siteChat *sitechatsvc.Commands, hub *Hub) {
	sitechathttp.MountPublicRoutes(r, siteChat, hub)
}

func MountOperatorHTTP(r chi.Router, siteChat *sitechatsvc.Commands, hub *Hub) {
	sitechathttp.MountOperatorRoutes(r, siteChat, hub)
}

func MountProjectHTTP(r chi.Router, siteChat *sitechatsvc.Commands) {
	sitechathttp.MountProjectRoutes(r, siteChat)
}
