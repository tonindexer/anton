package http

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
	"github.com/tonindexer/anton/internal/core/filter"
)

// @title      		Anton
// @version     	0.1
// @description 	Project fetches data from TON blockchain.

// @contact.name   	Anton
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

func paramErr(c *gin.Context, param string, err error) {
	c.IndentedJSON(http.StatusBadRequest, gin.H{"param": param, "error": err.Error()})
}

func internalErr(c *gin.Context, err error) {
	if errors.Is(err, core.ErrInvalidArg) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Error().Err(err).
		Str("path", c.FullPath()).
		Str("url", c.Request.URL.String()).
		Msg("internal server error")

	c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func unmarshalOperationID(op string) (uint32, error) {
	op = strings.Replace(op, "0x", "", 1)

	i, err := strconv.ParseInt(op, 10, 64)
	if err == nil {
		return uint32(i), nil //nolint:gosec // no integer overflow
	}

	i, err = strconv.ParseInt(op, 16, 64)
	if err == nil {
		return uint32(i), nil //nolint:gosec // no integer overflow
	}

	return 0, err
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
//
//	@Summary		statistics on all tables
//	@Description	Returns statistics on blocks, transactions, messages and accounts
//	@Tags			statistics
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}		aggregate.Statistics
//	@Router			/statistics [get]
func (ctrl *Controller) GetStatistics(c *gin.Context) {
	ret, err := ctrl.svc.GetStatistics(c.Request.Context())
	if err != nil {
		internalErr(c, err)
		return
	}
	c.IndentedJSON(http.StatusOK, ret)
}

type GetInterfacesRes struct {
	Total   int                       `json:"total"`
	Results []*core.ContractInterface `json:"results"`
}

// GetInterfaces godoc
//
//	@Summary		contract interfaces
//	@Description	Returns known contract interfaces
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}		GetInterfacesRes
//	@Router			/contracts/interfaces [get]
func (ctrl *Controller) GetInterfaces(c *gin.Context) {
	ret, err := ctrl.svc.GetInterfaces(c.Request.Context())
	if err != nil {
		internalErr(c, err)
		return
	}
	c.IndentedJSON(http.StatusOK, GetInterfacesRes{Total: len(ret), Results: ret})
}

type GetOperationsRes struct {
	Total   int                       `json:"total"`
	Results []*core.ContractOperation `json:"results"`
}

// GetOperations godoc
//
//	@Summary		contract operations
//	@Description	Returns known contract message payloads schema
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}		GetOperationsRes
//	@Router			/contracts/operations [get]
func (ctrl *Controller) GetOperations(c *gin.Context) {
	ret, err := ctrl.svc.GetOperations(c.Request.Context())
	if err != nil {
		internalErr(c, err)
		return
	}
	c.IndentedJSON(http.StatusOK, GetOperationsRes{Total: len(ret), Results: ret})
}

type GetDefinitionsRes struct {
	Total   int                               `json:"total"`
	Results map[abi.TLBType]abi.TLBFieldsDesc `json:"results"`
}

// GetDefinitions godoc
//
//	@Summary		struct definitions
//	@Description	Returns definitions used in messages and get-methods parsing
//	@Tags			contract
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}		GetDefinitionsRes
//	@Router			/contracts/definitions [get]
func (ctrl *Controller) GetDefinitions(c *gin.Context) {
	ret, err := ctrl.svc.GetDefinitions(c.Request.Context())
	if err != nil {
		internalErr(c, err)
		return
	}
	c.IndentedJSON(http.StatusOK, GetDefinitionsRes{Total: len(ret), Results: ret})
}

// GetBlocks godoc
//
//	@Summary		block info
//	@Description	Returns filtered blocks
//	@Tags			block
//	@Accept			json
//	@Produce		json
//	@Param   		workchain     		query   int 	false   "workchain"						default(-1)
//	@Param   		shard	     		query   int64 	false   "shard"
//	@Param   		seq_no	     		query   int 	false   "seq_no"
//	@Param   		with_transactions	query	bool  	false	"include transactions"			default(false)
//	@Param			order				query	string	false	"order by seq_no"				Enums(ASC, DESC) default(DESC)
//	@Param   		after	     		query   int 	false	"start from this seq_no"
//	@Param   		limit	     		query   int 	false	"limit"							default(3) maximum(100)
//	@Param   		count	     		query   bool 	false	"count total number of rows"	default(false)
//	@Success		200		{object}	filter.BlocksRes
//	@Router			/blocks [get]
func (ctrl *Controller) GetBlocks(c *gin.Context) {
	var req filter.BlocksReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "block_filter", err)
		return
	}
	if req.Limit > 100 {
		paramErr(c, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	if mw := int32(-1); c.Query("workchain") == "" {
		req.Workchain = &mw
	}

	req.WithShards = true
	if req.WithTransactions {
		req.WithTransactions = true
		req.WithTransactionAccountState = true
		req.WithTransactionMessages = true
	}

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(c, "order", err)
		return
	}

	ret, err := ctrl.svc.FilterBlocks(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

type GetLabelCategoriesRes struct {
	Total   int                  `json:"total"`
	Results []core.LabelCategory `json:"results"`
}

// GetLabelCategories godoc
//
//	@Summary		address label categories
//	@Description	Returns all possible label categories
//	@Tags			label
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	GetLabelCategoriesRes
//	@Router			/labels/categories [get]
func (ctrl *Controller) GetLabelCategories(c *gin.Context) {
	ret, err := ctrl.svc.GetLabelCategories(c.Request.Context())
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, GetLabelCategoriesRes{Total: len(ret), Results: ret})
}

// GetLabels godoc
//
//	@Summary		address labels
//	@Description	Search addresses by label name or category
//	@Tags			label
//	@Accept			json
//	@Produce		json
//	@Param   		name				query	string  	false	"filter labels by its name"
//	@Param   		category			query	[]string  	false	"filter by categories"
//	@Param   		offset	     		query   int 		false	"offset"
//	@Param   		limit	     		query   int 		false	"limit"										default(3) maximum(10000)
//	@Success		200		{object}	filter.LabelsRes
//	@Router			/labels [get]
func (ctrl *Controller) GetLabels(c *gin.Context) {
	var req filter.LabelsReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "label_filter", err)
		return
	}
	if req.Limit > 10000 {
		paramErr(c, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	ret, err := ctrl.svc.FilterLabels(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

// GetAccounts godoc
//
//	@Summary		account data
//	@Description	Returns account states and its parsed data
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Param   		address     		query   []string 	false   "only given addresses"
//	@Param   		latest				query	bool  		false	"only latest account states"
//	@Param   		interface			query	[]string  	false	"filter by interfaces"
//	@Param   		owner_address		query	string  	false	"filter FT wallets or NFT items by owner address"
//	@Param   		minter_address		query	string  	false	"filter FT wallets or NFT items by minter address"
//	@Param			order				query	string		false	"order by last_tx_lt"						Enums(ASC, DESC) default(DESC)
//	@Param   		after	     		query   int 		false	"start from this last_tx_lt"
//	@Param   		limit	     		query   int 		false	"limit"										default(3) maximum(10000)
//	@Param   		count	     		query   bool 		false	"count total number of rows"				default(false)
//	@Success		200		{object}	filter.AccountsRes
//	@Router			/accounts [get]
func (ctrl *Controller) GetAccounts(c *gin.Context) {
	req := filter.AccountsReq{WithCodeData: true}

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "account_filter", err)
		return
	}
	if req.Limit > 10000 {
		paramErr(c, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Addresses, err = getAddresses(c, "address")
	if err != nil {
		paramErr(c, "address", err)
		return
	}
	req.OwnerAddress, err = unmarshalAddress(c.Query("owner_address"))
	if err != nil {
		paramErr(c, "owner_address", err)
		return
	}
	req.MinterAddress, err = unmarshalAddress(c.Query("minter_address"))
	if err != nil {
		paramErr(c, "minter_address", err)
		return
	}

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(c, "order", err)
		return
	}

	ret, err := ctrl.svc.FilterAccounts(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

// AggregateAccounts godoc
//
//	@Summary		aggregated account data
//	@Description	Aggregates FT or NFT data filtered by minter address
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Param   		address				query	string  	false	"address on which statistics are calculated"
//	@Param   		minter_address		query	string  	false	"NFT collection or FT master address"
//	@Param   		limit	     		query   int 		false	"limit"									default(25) maximum(1000000)
//	@Success		200		{object}	aggregate.AccountsRes
//	@Router			/accounts/aggregated [get]
func (ctrl *Controller) AggregateAccounts(c *gin.Context) {
	var req aggregate.AccountsReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "account_filter", err)
		return
	}
	if req.Limit > 1000000 {
		paramErr(c, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Address, err = unmarshalAddress(c.Query("address"))
	if err != nil {
		paramErr(c, "address", err)
		return
	}

	req.MinterAddress, err = unmarshalAddress(c.Query("minter_address"))
	if err != nil {
		paramErr(c, "minter_address", err)
		return
	}

	ret, err := ctrl.svc.AggregateAccounts(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

// AggregateAccountsHistory godoc
//
//	@Summary		aggregated accounts grouped by timestamp
//	@Description	Counts accounts
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Param   		metric				query	string  	true	"metric to show"			Enums(active_addresses)
//	@Param   		interface			query	[]string  	false	"filter by interfaces"
//	@Param   		minter_address		query	string  	false	"NFT collection or FT master address"
//	@Param   		from				query	string  	false	"from timestamp"
//	@Param   		to					query	string  	false	"to timestamp"
//	@Param   		interval			query	string  	true	"group interval"			Enums(24h, 8h, 4h, 1h, 15m)
//	@Success		200		{object}	history.AccountsRes
//	@Router			/accounts/aggregated/history [get]
func (ctrl *Controller) AggregateAccountsHistory(c *gin.Context) {
	var req history.AccountsReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "account_filter", err)
		return
	}

	req.MinterAddress, err = unmarshalAddress(c.Query("minter_address"))
	if err != nil {
		paramErr(c, "minter_address", err)
		return
	}

	ret, err := ctrl.svc.AggregateAccountsHistory(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

// GetTransactions godoc
//
//	@Summary		transactions data
//	@Description	Returns transactions, states and messages
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//	@Param   		address     		query   []string 	false   "only given addresses"
//	@Param   		hash				query	string  	false	"search by tx hash"
//	@Param   		in_msg_hash			query	string  	false	"search by incoming message hash"
//	@Param   		workchain			query	int32  		false	"filter by workchain"
//	@Param			created_lt			query	uint64		false	"search by created_lt"
//	@Param			order				query	string		false	"order by created_lt"			Enums(ASC, DESC) default(DESC)
//	@Param   		after	     		query   int 		false	"start from this created_lt"
//	@Param   		limit	     		query   int 		false	"limit"							default(3) maximum(10000)
//	@Param   		count	     		query   bool 		false	"count total number of rows"	default(false)
//	@Success		200		{object}	filter.TransactionsRes
//	@Router			/transactions [get]
func (ctrl *Controller) GetTransactions(c *gin.Context) {
	var req filter.TransactionsReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "tx_filter", err)
		return
	}
	if req.Limit > 10000 {
		paramErr(c, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Hash, err = unmarshalBytes(c.Query("hash"))
	if err != nil {
		paramErr(c, "hash", err)
		return
	}
	req.InMsgHash, err = unmarshalBytes(c.Query("in_msg_hash"))
	if err != nil {
		paramErr(c, "in_msg_hash", err)
		return
	}

	req.WithAccountState = true
	req.WithMessages = true

	req.Addresses, err = getAddresses(c, "address")
	if err != nil {
		paramErr(c, "address", err)
		return
	}

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(c, "order", err)
		return
	}

	ret, err := ctrl.svc.FilterTransactions(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}
	c.IndentedJSON(http.StatusOK, ret)
}

// AggregateTransactionsHistory godoc
//
//	@Summary		aggregated transactions grouped by timestamp
//	@Description	Counts transactions
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//	@Param   		metric				query	string  	true	"metric to show"			Enums(transaction_count)
//	@Param   		address     		query   []string 	false   "tx address"
//	@Param   		workchain     		query  	int32  		false	"filter by workchain"
//	@Param   		from				query	string  	false	"from timestamp"
//	@Param   		to					query	string  	false	"to timestamp"
//	@Param   		interval			query	string  	true	"group interval"			Enums(24h, 8h, 4h, 1h, 15m)
//	@Success		200		{object}	history.TransactionsRes
//	@Router			/transactions/aggregated/history [get]
func (ctrl *Controller) AggregateTransactionsHistory(c *gin.Context) {
	var req history.TransactionsReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "tx_filter", err)
		return
	}

	req.Addresses, err = getAddresses(c, "address")
	if err != nil {
		paramErr(c, "address", err)
		return
	}

	ret, err := ctrl.svc.AggregateTransactionsHistory(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

// GetMessages godoc
//
//	@Summary		transaction messages
//	@Description	Returns filtered messages
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//	@Param   		hash				query	string  	false	"msg hash"
//	@Param   		src_workchain     	query  	int32  		false	"filter by source workchain"
//	@Param   		dst_workchain     	query  	int32  		false	"filter by destination workchain"
//	@Param   		src_address     	query   []string 	false   "source address"
//	@Param   		dst_address     	query   []string 	false   "destination address"
//	@Param   		operation_id     	query   string 		false   "operation id in hex format or as int32"
//	@Param   		src_contract		query	[]string  	false	"source contract interface"
//	@Param   		dst_contract		query	[]string  	false	"destination contract interface"
//	@Param   		operation_name		query	[]string  	false	"filter by contract operation names"
//	@Param			order				query	string		false	"order by created_lt"						Enums(ASC, DESC) default(DESC)
//	@Param   		after	     		query   int 		false	"start from this created_lt"
//	@Param   		limit	     		query   int 		false	"limit"										default(3) maximum(10000)
//	@Param   		count	     		query   bool 		false	"count total number of rows"				default(false)
//	@Success		200		{object}	filter.MessagesRes
//	@Router			/messages [get]
func (ctrl *Controller) GetMessages(c *gin.Context) {
	var req filter.MessagesReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "msg_filter", err)
		return
	}
	if req.Limit > 10000 {
		paramErr(c, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Hash, err = unmarshalBytes(c.Query("hash"))
	if err != nil {
		paramErr(c, "hash", err)
		return
	}
	req.SrcAddresses, err = getAddresses(c, "src_address")
	if err != nil {
		paramErr(c, "src_address", err)
		return
	}
	req.DstAddresses, err = getAddresses(c, "dst_address")
	if err != nil {
		paramErr(c, "dst_address", err)
		return
	}

	if op := c.Query("operation_id"); op != "" {
		id, err := unmarshalOperationID(op)
		if err != nil {
			paramErr(c, "operation_id", err)
			return
		}
		req.OperationID = &id
	}

	req.Order, err = unmarshalSorting(req.Order)
	if err != nil {
		paramErr(c, "order", err)
		return
	}

	ret, err := ctrl.svc.FilterMessages(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}
	c.IndentedJSON(http.StatusOK, ret)
}

// AggregateMessages godoc
//
//	@Summary		aggregated messages
//	@Description	Aggregates receivers and senders
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//	@Param   		address				query	string  	true	"address to aggregate by"
//	@Param   		order_by	     	query   string 		true	"order aggregated by amount or message count"	Enums(amount, count)	default(amount)
//	@Param   		from				query	string  	false	"from timestamp"
//	@Param   		to					query	string  	false	"to timestamp"
//	@Param   		limit	     		query   int 		false	"limit"											default(25) maximum(1000000)
//	@Success		200		{object}	aggregate.MessagesRes
//	@Router			/messages/aggregated [get]
func (ctrl *Controller) AggregateMessages(c *gin.Context) {
	var req aggregate.MessagesReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "msg_filter", err)
		return
	}
	if req.Limit > 1000000 {
		paramErr(c, "limit", errors.Wrapf(core.ErrInvalidArg, "limit is too big"))
		return
	}

	req.Address, err = unmarshalAddress(c.Query("address"))
	if err != nil {
		paramErr(c, "address", err)
		return
	}

	switch req.OrderBy {
	case "amount", "count":
	default:
		paramErr(c, "order_by", errors.Wrap(core.ErrInvalidArg, "wrong order_by argument"))
		return
	}

	ret, err := ctrl.svc.AggregateMessages(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

// AggregateMessagesHistory godoc
//
//	@Summary		aggregated messages grouped by timestamp
//	@Description	Counts messages or sums amount
//	@Tags			transaction
//	@Accept			json
//	@Produce		json
//	@Param   		metric				query	string  	true	"metric to show"								Enums(message_count, message_amount_sum)
//	@Param   		src_address     	query   []string 	false   "source address"
//	@Param   		dst_address     	query   []string 	false   "destination address"
//	@Param   		src_workchain     	query  	int32  		false	"source workchain"
//	@Param   		dst_workchain     	query  	int32  		false	"destination workchain"
//	@Param   		src_contract		query	[]string  	false	"source contract interface"
//	@Param   		dst_contract		query	[]string  	false	"destination contract interface"
//	@Param   		operation_name		query	[]string  	false	"contract operation names"
//	@Param   		minter_address		query	string  	false	"filter FT or NFT operations by minter address"
//	@Param   		from				query	string  	false	"from timestamp"
//	@Param   		to					query	string  	false	"to timestamp"
//	@Param   		interval			query	string  	true	"group interval"								Enums(24h, 8h, 4h, 1h, 15m)
//	@Success		200		{object}	history.MessagesRes
//	@Router			/messages/aggregated/history [get]
func (ctrl *Controller) AggregateMessagesHistory(c *gin.Context) {
	var req history.MessagesReq

	err := c.ShouldBindQuery(&req)
	if err != nil {
		paramErr(c, "msg_filter", err)
		return
	}

	req.SrcAddresses, err = getAddresses(c, "src_address")
	if err != nil {
		paramErr(c, "src_address", err)
		return
	}
	req.DstAddresses, err = getAddresses(c, "dst_address")
	if err != nil {
		paramErr(c, "dst_address", err)
		return
	}
	req.MinterAddress, err = unmarshalAddress(c.Query("minter_address"))
	if err != nil {
		paramErr(c, "minter_address", err)
		return
	}

	ret, err := ctrl.svc.AggregateMessagesHistory(c.Request.Context(), &req)
	if err != nil {
		internalErr(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}
