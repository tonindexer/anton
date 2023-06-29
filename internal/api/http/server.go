package http

import (
	"net/http"

	"github.com/gin-contrib/cors"

	_ "github.com/tonindexer/anton/api/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

type QueryController interface {
	GetStatistics(*gin.Context)

	GetBlocks(*gin.Context)

	GetLabelCategories(*gin.Context)
	GetLabels(*gin.Context)

	GetAccounts(*gin.Context)
	AggregateAccounts(*gin.Context)
	AggregateAccountsHistory(*gin.Context)

	GetTransactions(*gin.Context)
	AggregateTransactionsHistory(*gin.Context)

	GetMessages(*gin.Context)
	AggregateMessages(*gin.Context)
	AggregateMessagesHistory(*gin.Context)

	GetInterfaces(*gin.Context)
	GetOperations(*gin.Context)
}

type Server struct {
	listenHost string
	router     *gin.Engine
}

func NewServer(host string) *Server {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))
	return &Server{listenHost: host, router: router}
}

func (s *Server) RegisterRoutes(t QueryController) {
	base := s.router.Group(basePath)

	base.GET("/statistics", t.GetStatistics)

	base.GET("/blocks", t.GetBlocks)

	base.GET("/labels", t.GetLabels)
	base.GET("/labels/categories", t.GetLabelCategories)

	base.GET("/accounts", t.GetAccounts)
	base.GET("/accounts/aggregated", t.AggregateAccounts)
	base.GET("/accounts/aggregated/history", t.AggregateAccountsHistory)

	base.GET("/transactions", t.GetTransactions)
	base.GET("/transactions/aggregated/history", t.AggregateTransactionsHistory)

	base.GET("/messages", t.GetMessages)
	base.GET("/messages/aggregated", t.AggregateMessages)
	base.GET("/messages/aggregated/history", t.AggregateMessagesHistory)

	base.GET("/contract/interfaces", t.GetInterfaces)
	base.GET("/contract/operations", t.GetOperations)

	base.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL(basePath+"/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1)))

	base.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, basePath+"/swagger/index.html")
	})
}

func (s *Server) Run() error {
	return s.router.Run(s.listenHost)
}
