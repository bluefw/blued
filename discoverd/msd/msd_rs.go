package msd

import (
	"encoding/base64"
	"github.com/bluefw/blued/discoverd/api"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type ServiceResource struct {
	repo   *DiscoverdRepo
	logger *log.Logger
}

func NewServiceResource(dr *DiscoverdRepo, l *log.Logger) *ServiceResource {
	return &ServiceResource{
		repo:   dr,
		logger: l,
	}
}

func (sr *ServiceResource) RegMicroApp(c *gin.Context) {
	var as api.MicroApp
	if err := c.Bind(&as); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("problem decoding body"))
		return
	}

	sr.repo.Register(&as)
	c.JSON(http.StatusCreated, nil)
}

func (sr *ServiceResource) GetRouterTable(c *gin.Context) {
	addr, err := base64.StdEncoding.DecodeString(c.Params.ByName("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("error decoding addr"))
		return
	}
	rt := sr.repo.GetRouterTable(string(addr))
	c.JSON(http.StatusOK, rt)
}

func (sr *ServiceResource) Refresh(c *gin.Context) {
	addr, err := base64.StdEncoding.DecodeString(c.Params.ByName("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("error decoding addr"))
		return
	}
	appStatus := sr.repo.Refresh(string(addr))
	c.JSON(http.StatusAccepted, appStatus)
}
