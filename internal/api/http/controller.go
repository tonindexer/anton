package http

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
	"github.com/tonindexer/anton/internal/core/filter"
)

// @title      		tonidx
// @version     	0.1
// @description 	Project fetches data from TON blockchain.

// @contact.name   	Dat Boi
// @contact.url    	https://anton.tools

// @license.name  	Apache 2.0
// @license.url   	http://www.apache.org/licenses/LICENSE-2.0.html

// @host      		anton.tools
// @BasePath  		/api/v0
// @schemes 		https

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
	if errors.Is(err, core.ErrInvalidArg) {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Error().Str("path", ctx.FullPath()).Err(err).Msg("internal server error")
	ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func unmarshalAddress(a string) (*addr.Address, error) {
	if a == "" {
		return nil, nil //nolint:nilnil // lazy ...
	}

	var x = new(addr.Address)

	if err := x.UnmarshalJSON([]byte(a)); err != nil {
		return nil, errors.Wrapf(core.ErrInvalidArg, "unmarshal %s address (%s)", a, err.Error())
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
		return "", errors.Wrap(core.ErrInvalidArg, "only DESC and ASC sorting available")
	}
}

func unmarshalBytes(x string) ([]byte, error) {
	if x == "" {
		return nil, nil
	}
	if ret, err := hex.DecodeString(x); err == nil {
		return ret, nil
	}
	if ret, err := base64.StdEncoding.DecodeString(x); err == nil {
		return ret, nil
	}
	return nil, errors.Wrapf(core.ErrInvalidArg, "cannot decode bytes %s", x)
}

func getAddresses(ctx *gin.Context, name string) ([]*addr.Address, error) {
	var ret []*addr.Address

	for _, a := range ctx.Request.URL.Query()[name] {
		x, err := unmarshalAddress(a)
		if err != nil {
			return nil, err
		}
		ret = append(ret, x)
	}

	return ret, nil
}

// GetStatistics godoc
//	@Summary		statistics on all tables
//	@Description	Returns statistics on blocks, transactions, messages and accounts
//	@Tags			statistics
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}		aggregate.Statistics
//	@Router			/statistics [get]
func (c *Controller) GetStatistics(ctx *gin.Context) {
	ret, err := c.svc.GetStatistics(ctx)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

type GetInterfacesRes struct {
	Total   int                       `json:"total"`
	Results []*core.ContractInterface `json:"results"`
}

// GetInterfaces godoc
//	@Summary		contract interfaces
//	@Description	Returns known contract interfaces
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}		GetInterfacesRes
//	@Router			/contract/interfaces [get]
func (c *Controller) GetInterfaces(ctx *gin.Context) {
	ret, err := c.svc.GetInterfaces(ctx)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, GetInterfacesRes{Total: len(ret), Results: ret})
}

type GetOperationsRes struct {
	Total   int                       `json:"total"`
	Results []*core.ContractOperation `json:"results"`
}

// GetOperations godoc
//	@Summary		contract operations
//	@Description	Returns known contract message payloads schema
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}		GetOperationsRes
//	@Router			/contract/operations [get]
func (c *Controller) GetOperations(ctx *gin.Context) {
	ret, err := c.svc.GetOperations(ctx)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, GetOperationsRes{Total: len(ret), Results: ret})
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
//  @Param   		limit	     		query   int 	false	"limit"						default(3) maximum(100)
//	@Success		200		{object}	filter.BlocksRes
//	@Router			/blocks [get]
func (c *Controller) GetBlocks(ctx *gin.Context) {
	var req filter.BlocksReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "block_filter", err)
		return
	}
	if req.Limit > 100 {
		paramErr(ctx, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	if mw := int32(-1); ctx.Query("workchain") == "" {
		req.Workchain = &mw
	}

	req.WithShards = true
	if req.WithTransactions {
		req.WithTransactions = true
		req.WithTransactionAccountState = true
		req.WithTransactionAccountData = true
		req.WithTransactionMessages = true
		req.WithTransactionMessagePayloads = true
	}

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.FilterBlocks(ctx, &req)
	if err != nil {
		internalErr(ctx, err)
		return
	}

	ctx.IndentedJSON(http.StatusOK, ret)
}

// GetAccounts godoc
//	@Summary		account data
//	@Description	Returns account states and its parsed data
//	@Tags			account
//	@Accept			json
//	@Produce		json
//  @Param   		address     		query   []string 	false   "only given addresses"
//  @Param   		latest				query	bool  		false	"only latest account states"
//  @Param   		interface			query	[]string  	false	"filter by interfaces"
//  @Param   		owner_address		query	string  	false	"filter FT wallets or NFT items by owner address"
//  @Param   		minter_address		query	string  	false	"filter FT wallets or NFT items by minter address"
//  @Param			order				query	string		false	"order by last_tx_lt"						Enums(ASC, DESC) default(DESC)
//  @Param   		after	     		query   int 		false	"start from this last_tx_lt"
//  @Param   		limit	     		query   int 		false	"limit"										default(3) maximum(10000)
//	@Success		200		{object}	filter.AccountsRes
//	@Router			/accounts [get]
func (c *Controller) GetAccounts(ctx *gin.Context) {
	var req filter.AccountsReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "account_filter", err)
		return
	}
	if req.Limit > 10000 {
		paramErr(ctx, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.WithData = true

	req.Addresses, err = getAddresses(ctx, "address")
	if err != nil {
		paramErr(ctx, "address", err)
		return
	}
	req.OwnerAddress, err = unmarshalAddress(ctx.Query("owner_address"))
	if err != nil {
		paramErr(ctx, "owner_address", err)
		return
	}
	req.MinterAddress, err = unmarshalAddress(ctx.Query("minter_address"))
	if err != nil {
		paramErr(ctx, "minter_address", err)
		return
	}

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.FilterAccounts(ctx, &req)
	if err != nil {
		internalErr(ctx, err)
		return
	}

	ctx.IndentedJSON(http.StatusOK, ret)
}

// AggregateAccounts godoc
//	@Summary		aggregated account data
//	@Description	Aggregates FT or NFT data filtered by minter address
//	@Tags			account
//	@Accept			json
//	@Produce		json
//  @Param   		minter_address		query	string  	true	"NFT collection or FT master address"
//  @Param   		limit	     		query   int 		false	"limit"									default(25) maximum(1000000)
//	@Success		200		{object}	aggregate.AccountsRes
//	@Router			/accounts/aggregated [get]
func (c *Controller) AggregateAccounts(ctx *gin.Context) {
	var req aggregate.AccountsReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "account_filter", err)
		return
	}
	if req.Limit > 1000000 {
		paramErr(ctx, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.MinterAddress, err = unmarshalAddress(ctx.Query("minter_address"))
	if err != nil {
		paramErr(ctx, "minter_address", err)
		return
	}

	ret, err := c.svc.AggregateAccounts(ctx, &req)
	if err != nil {
		internalErr(ctx, err)
		return
	}

	ctx.IndentedJSON(http.StatusOK, ret)
}

// AggregateAccountsHistory godoc
//	@Summary		aggregated accounts grouped by timestamp
//	@Description	Counts accounts
//	@Tags			account
//	@Accept			json
//	@Produce		json
//  @Param   		metric				query	string  	true	"metric to show"			Enums(active_addresses)
//  @Param   		interface			query	[]string  	false	"filter by interfaces"
//  @Param   		minter_address		query	string  	false	"NFT collection or FT master address"
//  @Param   		from				query	string  	false	"from timestamp"
//  @Param   		to					query	string  	false	"to timestamp"
//  @Param   		interval			query	string  	true	"group interval"			Enums(24h, 8h, 4h, 1h, 15m)
//	@Success		200		{object}	history.AccountsRes
//	@Router			/accounts/aggregated/history [get]
func (c *Controller) AggregateAccountsHistory(ctx *gin.Context) {
	var req history.AccountsReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "account_filter", err)
		return
	}

	req.MinterAddress, err = unmarshalAddress(ctx.Query("minter_address"))
	if err != nil {
		paramErr(ctx, "minter_address", err)
		return
	}

	ret, err := c.svc.AggregateAccountsHistory(ctx, &req)
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
//  @Param   		address     		query   []string 	false   "only given addresses"
//  @Param   		hash				query	string  	false	"search by tx hash"
//  @Param   		in_msg_hash			query	string  	false	"search by incoming message hash"
//  @Param   		workchain			query	int32  		false	"filter by workchain"
//  @Param			order				query	string		false	"order by created_lt"			Enums(ASC, DESC) default(DESC)
//  @Param   		after	     		query   int 		false	"start from this created_lt"
//  @Param   		limit	     		query   int 		false	"limit"							default(3) maximum(10000)
//	@Success		200		{object}	filter.TransactionsRes
//	@Router			/transactions [get]
func (c *Controller) GetTransactions(ctx *gin.Context) {
	var req filter.TransactionsReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "tx_filter", err)
		return
	}
	if req.Limit > 10000 {
		paramErr(ctx, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Hash, err = unmarshalBytes(ctx.Query("hash"))
	if err != nil {
		paramErr(ctx, "hash", err)
		return
	}
	req.InMsgHash, err = unmarshalBytes(ctx.Query("in_msg_hash"))
	if err != nil {
		paramErr(ctx, "in_msg_hash", err)
		return
	}

	req.WithAccountState = true
	req.WithAccountData = true
	req.WithMessages = true
	req.WithMessagePayloads = true

	req.Addresses, err = getAddresses(ctx, "address")
	if err != nil {
		paramErr(ctx, "address", err)
		return
	}

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.FilterTransactions(ctx, &req)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

// AggregateTransactionsHistory godoc
//	@Summary		aggregated transactions grouped by timestamp
//	@Description	Counts transactions
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//  @Param   		metric				query	string  	true	"metric to show"			Enums(transaction_count)
//  @Param   		address     		query   []string 	false   "tx address"
//  @Param   		workchain     		query  	int32  		false	"filter by workchain"
//  @Param   		from				query	string  	false	"from timestamp"
//  @Param   		to					query	string  	false	"to timestamp"
//  @Param   		interval			query	string  	true	"group interval"			Enums(24h, 8h, 4h, 1h, 15m)
//	@Success		200		{object}	history.TransactionsRes
//	@Router			/transactions/aggregated/history [get]
func (c *Controller) AggregateTransactionsHistory(ctx *gin.Context) {
	var req history.TransactionsReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "tx_filter", err)
		return
	}

	req.Addresses, err = getAddresses(ctx, "address")
	if err != nil {
		paramErr(ctx, "address", err)
		return
	}

	ret, err := c.svc.AggregateTransactionsHistory(ctx, &req)
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
//  @Param   		src_address     	query   []string 	false   "source address"
//  @Param   		dst_address     	query   []string 	false   "destination address"
//  @Param   		src_contract		query	[]string  	false	"source contract interface"
//  @Param   		dst_contract		query	[]string  	false	"destination contract interface"
//  @Param   		operation_name		query	[]string  	false	"filter by contract operation names"
//  @Param   		minter_address		query	string  	false	"filter FT or NFT operations by minter address"
//  @Param			order				query	string		false	"order by created_lt"						Enums(ASC, DESC) default(DESC)
//  @Param   		after	     		query   int 		false	"start from this created_lt"
//  @Param   		limit	     		query   int 		false	"limit"										default(3) maximum(10000)
//	@Success		200		{object}	filter.MessagesRes
//	@Router			/messages [get]
func (c *Controller) GetMessages(ctx *gin.Context) {
	var req filter.MessagesReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "msg_filter", err)
		return
	}
	if req.Limit > 10000 {
		paramErr(ctx, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Hash, err = unmarshalBytes(ctx.Query("hash"))
	if err != nil {
		paramErr(ctx, "hash", err)
		return
	}
	req.SrcAddresses, err = getAddresses(ctx, "src_address")
	if err != nil {
		paramErr(ctx, "src_address", err)
		return
	}
	req.DstAddresses, err = getAddresses(ctx, "dst_address")
	if err != nil {
		paramErr(ctx, "dst_address", err)
		return
	}
	req.MinterAddress, err = unmarshalAddress(ctx.Query("minter_address"))
	if err != nil {
		paramErr(ctx, "minter_address", err)
		return
	}

	req.WithPayload = true

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(ctx, "order", err)
		return
	}

	ret, err := c.svc.FilterMessages(ctx, &req)
	if err != nil {
		internalErr(ctx, err)
		return
	}
	ctx.IndentedJSON(http.StatusOK, ret)
}

// AggregateMessages godoc
//	@Summary		aggregated messages
//	@Description	Aggregates receivers and senders
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//  @Param   		address				query	string  	true	"address to aggregate by"
//  @Param   		order_by	     	query   string 		true	"order aggregated by amount or message count"	Enums(amount, count)	default(amount)
//  @Param   		limit	     		query   int 		false	"limit"											default(25) maximum(1000000)
//	@Success		200		{object}	aggregate.MessagesRes
//	@Router			/messages/aggregated [get]
func (c *Controller) AggregateMessages(ctx *gin.Context) {
	var req aggregate.MessagesReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "msg_filter", err)
		return
	}
	if req.Limit > 1000000 {
		paramErr(ctx, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Address, err = unmarshalAddress(ctx.Query("address"))
	if err != nil {
		paramErr(ctx, "address", err)
		return
	}

	switch req.OrderBy {
	case "amount", "count":
	default:
		paramErr(ctx, "order_by", errors.Wrap(core.ErrInvalidArg, "wrong order_by argument"))
		return
	}

	ret, err := c.svc.AggregateMessages(ctx, &req)
	if err != nil {
		internalErr(ctx, err)
		return
	}

	ctx.IndentedJSON(http.StatusOK, ret)
}

// AggregateMessagesHistory godoc
//	@Summary		aggregated messages grouped by timestamp
//	@Description	Counts messages or sums amount
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//  @Param   		metric				query	string  	true	"metric to show"								Enums(message_count, message_amount_sum)
//  @Param   		src_address     	query   []string 	false   "source address"
//  @Param   		dst_address     	query   []string 	false   "destination address"
//  @Param   		src_contract		query	[]string  	false	"source contract interface"
//  @Param   		dst_contract		query	[]string  	false	"destination contract interface"
//  @Param   		operation_name		query	[]string  	false	"filter by contract operation names"
//  @Param   		minter_address		query	string  	false	"filter FT or NFT operations by minter address"
//  @Param   		from				query	string  	false	"from timestamp"
//  @Param   		to					query	string  	false	"to timestamp"
//  @Param   		interval			query	string  	true	"group interval"								Enums(24h, 8h, 4h, 1h, 15m)
//	@Success		200		{object}	history.MessagesRes
//	@Router			/messages/aggregated/history [get]
func (c *Controller) AggregateMessagesHistory(ctx *gin.Context) {
	var req history.MessagesReq

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		paramErr(ctx, "msg_filter", err)
		return
	}

	req.SrcAddresses, err = getAddresses(ctx, "src_address")
	if err != nil {
		paramErr(ctx, "src_address", err)
		return
	}
	req.DstAddresses, err = getAddresses(ctx, "dst_address")
	if err != nil {
		paramErr(ctx, "dst_address", err)
		return
	}
	req.MinterAddress, err = unmarshalAddress(ctx.Query("minter_address"))
	if err != nil {
		paramErr(ctx, "minter_address", err)
		return
	}

	ret, err := c.svc.AggregateMessagesHistory(ctx, &req)
	if err != nil {
		internalErr(ctx, err)
		return
	}

	ctx.IndentedJSON(http.StatusOK, ret)
}
