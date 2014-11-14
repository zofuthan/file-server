package handlers

import (
	"github.com/cloudfoundry-incubator/file-server/ccclient"
	"github.com/cloudfoundry-incubator/file-server/handlers/static"
	"github.com/cloudfoundry-incubator/file-server/handlers/upload_build_artifacts"
	"github.com/cloudfoundry-incubator/file-server/handlers/upload_droplet"
	"github.com/cloudfoundry-incubator/runtime-schema/router"
	"github.com/pivotal-golang/lager"
)

func New(staticDirectory string, uploader ccclient.Uploader, poller ccclient.Poller, logger lager.Logger) router.Handlers {
	staticRoute, _ := router.NewFileServerRoutes().RouteForHandler(router.FS_STATIC)

	return router.Handlers{
		router.FS_STATIC:                 static.New(staticDirectory, staticRoute.Path, logger),
		router.FS_UPLOAD_DROPLET:         upload_droplet.New(uploader, poller, logger),
		router.FS_UPLOAD_BUILD_ARTIFACTS: upload_build_artifacts.New(uploader, logger),
	}
}
