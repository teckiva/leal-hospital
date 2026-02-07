package app

import (
	"github.com/gin-gonic/gin"
	"github.com/leal-hospital/server/di"
)

// Module represents a feature module
type Module interface {
	Name() string
	Configure(container *di.Container)
	RegisterRoutes(router *gin.Engine, container *di.Container)
}
