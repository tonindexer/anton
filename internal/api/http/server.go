package http

import (
	_ "github.com/iam047801/tonidx/api/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

type QueryController interface {
	GetInterfaces(*gin.Context)
	GetOperations(*gin.Context)

	GetBlocks(*gin.Context)

	GetAccountStates(*gin.Context)

	GetTransactions(*gin.Context)
	GetMessages(*gin.Context)
}

type Server struct {
	listenHost string
	router     *gin.Engine
}

func NewServer(host string) *Server {
	return &Server{listenHost: host, router: gin.Default()}
}

func (s *Server) RegisterRoutes(t QueryController) {
	base := s.router.Group(basePath)

	base.GET("/contract/interface", t.GetInterfaces)
	base.GET("/contract/operation", t.GetOperations)

	base.GET("/block", t.GetBlocks)

	base.GET("/account", t.GetAccountStates)

	base.GET("/transaction", t.GetTransactions)
	base.GET("/message", t.GetMessages)

	base.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/api/v1/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1)))
}

func (s *Server) Run() error {
	return s.router.Run(s.listenHost)
}
