package http

import (
	"net/http"

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

	base.GET("/contract/interfaces", t.GetInterfaces)
	base.GET("/contract/operations", t.GetOperations)

	base.GET("/blocks", t.GetBlocks)

	base.GET("/accounts", t.GetAccountStates)

	base.GET("/transactions", t.GetTransactions)
	base.GET("/messages", t.GetMessages)

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1)))

	s.router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

func (s *Server) Run() error {
	return s.router.Run(s.listenHost)
}
