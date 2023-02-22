package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/iam047801/tonidx/internal/app"
	"github.com/iam047801/tonidx/internal/core"
)

// @title      		tonidx
// @version         0.0.1
// @description     Project fetches data from TON blockchain.

// @contact.name   	Dat Boi
// @contact.url    	https://datboi420.t.me

// @license.name  	Apache 2.0
// @license.url   	http://www.apache.org/licenses/LICENSE-2.0.html

// @host      		localhost
// @BasePath  		/api/v1
// @schemes 		http

var basePath = "/api/v1"

var _ QueryController = (*Controller)(nil)

type Controller struct {
	svc app.QueryService
}

func NewController(svc app.QueryService) *Controller {
	return &Controller{svc: svc}
}

func paramErr(ctx *gin.Context, param string, err error) {
	ctx.IndentedJSON(http.StatusBadRequest, gin.H{"param": param, "error": err.Error()})
}

func internalErr(ctx *gin.Context, err error) {
	log.Error().Str("path", ctx.FullPath()).Err(err).Msg("internal server error")
	ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// GetInterfaces godoc
//	@Summary		contract interfaces
//	@Description	Returns known contract interfaces
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}		core.ContractInterface
//	@Router			/contract/interface [get]
func (c *Controller) GetInterfaces(ctx *gin.Context) {
	ret, err := c.svc.GetInterfaces(ctx)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

// GetOperations godoc
//	@Summary		contract operations
//	@Description	Returns known contract message payloads schema
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}		core.ContractOperation
//	@Router			/contract/operation [get]
func (c *Controller) GetOperations(ctx *gin.Context) {
	ret, err := c.svc.GetOperations(ctx)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

func getOffsetLimit(ctx *gin.Context) (int, int, error) {
	o, err := strconv.ParseInt(ctx.Query("offset"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	l, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return int(o), int(l), nil
}

// GetBlocks godoc
//	@Summary		block info
//	@Description	Returns filtered blocks
//	@Tags			block
//	@Accept			json
//	@Produce		json
//  @Param   		workchain     		query   int 	false   "workchain"
//  @Param   		shard	     		query   int 	false   "shard"
//  @Param   		seq_no	     		query   int 	false   "seq_no"
//  @Param   		with_transactions	query	bool  	false	"include transactions"
//  @Param   		offset	     		query   int 	true	"offset"
//  @Param   		limit	     		query   int 	true	"limit"
//	@Success		200		{array}		core.Block
//	@Router			/block [get]
func (c *Controller) GetBlocks(ctx *gin.Context) {
	var filter core.BlockFilter

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		paramErr(ctx, "block_filter", err)
		return
	}

	filter.WithShards = true
	if filter.WithTransactions {
		filter.WithTransactions = true
		filter.WithTransactionAccountState = true
		filter.WithTransactionAccountData = true
		filter.WithTransactionMessages = true
		filter.WithTransactionMessagePayloads = true
	}

	offset, limit, err := getOffsetLimit(ctx)
	if err != nil {
		paramErr(ctx, "offset_limit", err)
		return
	}

	ret, err := c.svc.GetBlocks(ctx, &filter, offset, limit)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

// GetAccountStates godoc
//	@Summary		account data
//	@Description	Returns account states and its parsed data
//	@Tags			account
//	@Accept			json
//	@Produce		json
//  @Param   		address     		query   string 		false   "account address"
//  @Param   		latest				query	bool  		false	"only latest account states"
//  @Param   		with_data			query	bool  		false	"include parsed data"
//  @Param   		interfaces			query	[]string  	false	"filter by interfaces"
//  @Param   		owner_address		query	string  	false	"filter FTs or NFTs by owner address"
//  @Param   		collection_address	query	string  	false	"filter NFT items by collection address"
//  @Param   		offset	     		query   int 		true	"offset"
//  @Param   		limit	     		query   int 		true	"limit"
//	@Success		200		{array}		core.AccountState
//	@Router			/account [get]
func (c *Controller) GetAccountStates(ctx *gin.Context) {
	var filter core.AccountStateFilter

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		paramErr(ctx, "account_filter", err)
		return
	}

	offset, limit, err := getOffsetLimit(ctx)
	if err != nil {
		paramErr(ctx, "offset_limit", err)
		return
	}

	ret, err := c.svc.GetAccountStates(ctx, &filter, offset, limit)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

// GetTransactions godoc
//	@Summary		transactions data
//	@Description	Returns transactions, states and messages
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//  @Param   		address     		query   string 		false   "account address"
//  @Param   		hash				query	string  	false	"tx hash"
//  @Param   		with_accounts		query	bool  		false	"with accounts"
//  @Param   		interfaces			query	[]string  	false	"filter by interfaces"
//  @Param   		offset	     		query   int 		true	"offset"
//  @Param   		limit	     		query   int 		true	"limit"
//	@Success		200		{array}		core.AccountState
//	@Router			/transaction [get]
func (c *Controller) GetTransactions(ctx *gin.Context) {
	var filter core.TransactionFilter

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		paramErr(ctx, "tx_filter", err)
		return
	}

	offset, limit, err := getOffsetLimit(ctx)
	if err != nil {
		paramErr(ctx, "offset_limit", err)
		return
	}

	ret, err := c.svc.GetTransactions(ctx, &filter, offset, limit)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

// GetMessages godoc
//	@Summary		transaction messages
//	@Description	Returns filtered messages
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//  @Param   		hash				query	string  	false	"msg hash"
//  @Param   		src_address     	query   string 		false   "source address"
//  @Param   		dst_address     	query   string 		false   "destination address"
//  @Param   		src_contract		query	string  	false	"source contract interface"
//  @Param   		dst_contract		query	string  	false	"destination contract interface"
//  @Param   		operation_names		query	[]string  	false	"filter by contract operation names"
//  @Param   		offset	     		query   int 		true	"offset"
//  @Param   		limit	     		query   int 		true	"limit"
//	@Success		200		{array}		core.AccountState
//	@Router			/message [get]
func (c *Controller) GetMessages(ctx *gin.Context) {
	var filter core.MessageFilter

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		paramErr(ctx, "msg_filter", err)
		return
	}
	filter.WithPayload = true

	offset, limit, err := getOffsetLimit(ctx)
	if err != nil {
		paramErr(ctx, "offset_limit", err)
		return
	}

	ret, err := c.svc.GetMessages(ctx, &filter, offset, limit)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}
