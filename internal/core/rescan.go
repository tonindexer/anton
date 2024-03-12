package core

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

type RescanTaskType string

const (
	// AddInterface task filters all account states by suitable addresses, code, or get method hashes.
	// From these account states, extract (address, last_tx_lt) pairs,
	// execute get methods on these pairs, and update the account states with the newly parsed data.
	AddInterface RescanTaskType = "add_interface"

	// UpdInterface filters the account states by the already set contract name.
	// Again, collect (address, last_tx_lt) pairs, execute get methods,
	// update the account states with the parsed data.
	UpdInterface RescanTaskType = "upd_interface"

	// DelInterface does the same filtering as UpdInterface,
	// but it clears any previously parsed data.
	DelInterface RescanTaskType = "del_interface"

	// AddGetMethod task executes this method across all account states that were previously scanned
	// and clears all parsed data in account states lacking the new get method.
	AddGetMethod RescanTaskType = "add_get_method"

	// DelGetMethod task eliminates the execution of this get-method in all previously parsed account states.
	// Then, it includes all account states that match the contract interface description, minus the deleted get method.
	DelGetMethod RescanTaskType = "del_get_method"

	// UpdGetMethod task simply iterates through all parsed account states associated with the specified contract name
	// and re-execute the changed get method.
	UpdGetMethod RescanTaskType = "upd_get_method"

	// UpdOperation task parses contract messages.
	// It iterates through all messages with specified operation id,
	// directed to (or originating from, in the case of outgoing operations) the given contract
	// and adds the parsed data.
	UpdOperation RescanTaskType = "upd_operation"

	// DelOperation task is the same algorithm, as UpdOperation, but it removes the parsed data.
	DelOperation RescanTaskType = "del_operation"
)

type RescanTask struct {
	bun.BaseModel `bun:"table:rescan_tasks" json:"-"`

	ID       int            `bun:",pk,autoincrement"`
	Finished bool           `bun:"finished,notnull"`
	Type     RescanTaskType `bun:"type:rescan_task_type,notnull"`

	// contract being rescanned
	ContractName abi.ContractName   `bun:",notnull" json:"contract_name"`
	Contract     *ContractInterface `bun:"rel:has-one,join:contract_name=name" json:"contract_interface"`

	// for get-method update
	ChangedGetMethod string `json:"changed_get_method,omitempty"`

	// for operations
	MessageType MessageType        `json:"message_type,omitempty"`
	Outgoing    bool               `json:"outgoing,omitempty"` // if operation is going from contract
	OperationID uint32             `json:"operation_id,omitempty"`
	Operation   *ContractOperation `bun:"rel:has-one,join:contract_name=contract_name,join:outgoing=outgoing,join:operation_id=operation_id" json:"contract_operation"`

	// checkpoint
	LastAddress *addr.Address `bun:"type:bytea" json:"last_address"`
	LastTxLt    uint64        `bun:"type:bigint" json:"last_tx_lt"`
}

type RescanRepository interface {
	AddRescanTask(ctx context.Context, task *RescanTask) error
	GetUnfinishedRescanTask(context.Context) (bun.Tx, *RescanTask, error)
	SetRescanTask(context.Context, bun.Tx, *RescanTask) error
}
