package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/iam047801/tonidx/internal/addr"
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
// @BasePath  		/api/v0
// @schemes 		http

var basePath = "/api/v0"

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

func unmarshalAddress(a string) (*addr.Address, error) {
	if len(a) == 0 {
		return nil, nil
	}

	var x = new(addr.Address)

	if err := x.UnmarshalJSON([]byte(a)); err != nil {
		return nil, errors.Wrapf(err, "unmarshal %s address", a)
	}

	return x, nil
}

func unmarshalSorting(sort string) (string, error) {
	switch sort = strings.ToUpper(sort); sort {
	case "", "DESC":
		return "DESC", nil
	case "ASC":
		return sort, nil
	default:
		return "", fmt.Errorf("only DESC and ASC sorting available")
	}
}

func getAddresses(ctx *gin.Context) ([]*addr.Address, error) {
	var ret []*addr.Address

	for _, a := range ctx.Request.URL.Query()["addresses"] {
		x, err := unmarshalAddress(a)
		if err != nil {
			return nil, err
		}
		ret = append(ret, x)
	}

	return ret, nil
}

// GetInterfaces godoc
//	@Summary		contract interfaces
//	@Description	Returns known contract interfaces
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}		core.ContractInterface
//	@Router			/contract/interfaces [get]
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
//	@Router			/contract/operations [get]
func (c *Controller) GetOperations(ctx *gin.Context) {
	ret, err := c.svc.GetOperations(ctx)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

// GetBlocks godoc
//	@Summary		block info
//	@Description	Returns filtered blocks
//	@Tags			block
//	@Accept			json
//	@Produce		json
//  @Param   		workchain     		query   int 	false   "workchain"					default(-1)
//  @Param   		shard	     		query   int64 	false   "shard"
//  @Param   		seq_no	     		query   int 	false   "seq_no"
//  @Param   		with_transactions	query	bool  	false	"include transactions"		default(false)
//  @Param			order				query	string	false	"order by seq_no"			Enums(ASC, DESC) default(DESC)
//  @Param   		after	     		query   int 	false	"start from this seq_no"
//  @Param   		limit	     		query   int 	false	"limit"						default(3)
//	@Success		200		{array}		core.Block
//	@Router			/blocks [get]
func (c *Controller) GetBlocks(ctx *gin.Context) {
	var filter core.BlockFilter

	err := ctx.ShouldBindQuery(&filter)
	if err != nil {
		paramErr(ctx, "block_filter", err)
		return
	}

	if mw := int32(-1); ctx.Query("workchain") == "" {
		filter.Workchain = &mw
	}

	filter.WithShards = true
	if filter.WithTransactions {
		filter.WithTransactions = true
		filter.WithTransactionAccountState = true
		filter.WithTransactionAccountData = true
		filter.WithTransactionMessages = true
		filter.WithTransactionMessagePayloads = true
	}

	filter.Order, err = unmarshalSorting(filter.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.GetBlocks(ctx, &filter)
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
//  @Param   		addresses     		query   []string 	false   "only given addresses"
//  @Param   		latest				query	bool  		false	"only latest account states"
//  @Param   		interfaces			query	[]string  	false	"filter by interfaces"
//  @Param   		owner_address		query	string  	false	"filter FTs or NFTs by owner address"
//  @Param   		collection_address	query	string  	false	"filter NFT items by collection address"
//  @Param   		master_address		query	string  	false	"filter FT wallets by minter address"
//  @Param			order				query	string		false	"order by last_tx_lt"						Enums(ASC, DESC) default(DESC)
//  @Param   		after	     		query   int 		false	"start from this last_tx_lt"
//  @Param   		limit	     		query   int 		false	"limit"										default(3)
//	@Success		200		{array}		core.AccountState
//	@Router			/accounts [get]
func (c *Controller) GetAccountStates(ctx *gin.Context) {
	var filter core.AccountStateFilter

	err := ctx.ShouldBindQuery(&filter)
	if err != nil {
		paramErr(ctx, "account_filter", err)
		return
	}

	filter.WithData = true

	filter.Addresses, err = getAddresses(ctx)
	if err != nil {
		paramErr(ctx, "addresses", err)
		return
	}
	filter.OwnerAddress, err = unmarshalAddress(ctx.Query("owner_address"))
	if err != nil {
		paramErr(ctx, "owner_address", err)
		return
	}
	filter.CollectionAddress, err = unmarshalAddress(ctx.Query("collection_address"))
	if err != nil {
		paramErr(ctx, "collection_address", err)
		return
	}
	filter.MasterAddress, err = unmarshalAddress(ctx.Query("master_address"))
	if err != nil {
		paramErr(ctx, "master_address", err)
		return
	}

	filter.Order, err = unmarshalSorting(filter.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.GetAccountStates(ctx, &filter)
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
//  @Param   		addresses     		query   []string 	false   "only given addresses"
//  @Param   		hash				query	string  	false	"search by tx hash"
//  @Param   		with_accounts		query	bool  		false	"return account states"
//  @Param   		interfaces			query	[]string  	false	"filter by interfaces"
//  @Param			order				query	string		false	"order by created_lt"			Enums(ASC, DESC) default(DESC)
//  @Param   		after	     		query   int 		false	"start from this created_lt"
//  @Param   		limit	     		query   int 		false	"limit"							default(3)
//	@Success		200		{array}		core.Transaction
//	@Router			/transactions [get]
func (c *Controller) GetTransactions(ctx *gin.Context) {
	var filter core.TransactionFilter

	err := ctx.ShouldBindQuery(&filter)
	if err != nil {
		paramErr(ctx, "tx_filter", err)
		return
	}

	filter.Addresses, err = getAddresses(ctx)
	if err != nil {
		paramErr(ctx, "addresses", err)
		return
	}

	filter.Order, err = unmarshalSorting(filter.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.GetTransactions(ctx, &filter)
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
//  @Param			order				query	string		false	"order by created_lt"						Enums(ASC, DESC) default(DESC)
//  @Param   		after	     		query   int 		false	"start from this created_lt"
//  @Param   		limit	     		query   int 		false	"limit"										default(3)
//	@Success		200		{array}		core.Message
//	@Router			/messages [get]
func (c *Controller) GetMessages(ctx *gin.Context) {
	var filter core.MessageFilter

	err := ctx.ShouldBindQuery(&filter)
	if err != nil {
		paramErr(ctx, "msg_filter", err)
		return
	}

	filter.WithPayload = true

	filter.Order, err = unmarshalSorting(filter.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.GetMessages(ctx, &filter)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}
