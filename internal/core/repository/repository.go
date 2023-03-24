package repository

import (
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/aggregate"
	"github.com/iam047801/tonidx/internal/core/filter"
)

type Block interface {
	core.BlockRepository
	filter.BlockRepository
}

type Account interface {
	core.AccountRepository
	filter.AccountRepository
	aggregate.AccountRepository
}

type Transaction interface {
	core.TransactionRepository
	filter.TransactionRepository
}

type Message interface {
	core.MessageRepository
	filter.MessageRepository
	aggregate.MessageRepository
}

type Contract interface {
	core.ContractRepository
}
