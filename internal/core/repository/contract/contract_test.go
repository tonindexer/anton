package contract_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	pg   *bun.DB
	repo *contract.Repository
)

func initdb(t testing.TB) {
	var (
		dsnPG = "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
		err   error
	)

	pg = bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG))), pgdialect.New())
	err = pg.Ping()
	require.Nil(t, err)

	repo = contract.NewRepository(pg)
}

func createTables(t testing.TB) {
	_, err := pg.ExecContext(context.Background(), "CREATE TYPE message_type AS ENUM (?, ?, ?)", core.ExternalIn, core.ExternalOut, core.Internal)
	require.Nil(t, err)
	err = contract.CreateTables(context.Background(), pg)
	require.Nil(t, err)
}

func dropTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := pg.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.ContractDefinition)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)

	_, err = pg.ExecContext(context.Background(), "DROP TYPE message_type")
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		t.Fatal(err)
	}
}

func TestRepository_AddContracts(t *testing.T) {
	initdb(t)

	d := core.ContractDefinition{
		Name: "test_definition",
	}

	definitionSchema := []byte(`[
  {
    "name": "new_owner",
    "tlb_type": "addr"
  },
  {
    "name": "response_destination",
    "tlb_type": "addr"
  },
  {
    "name": "custom_payload",
    "tlb_type": "maybe ^"
  },
  {
    "name": "forward_amount",
    "tlb_type": ".",
    "format": "coins"
  }
]`)
	err := json.Unmarshal(definitionSchema, &d.Schema)
	require.Nil(t, err)

	codeBoC, err := base64.StdEncoding.DecodeString("te6cckECFAEAAtQAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8GBwgJAubQAdDTAyFxsJJfBOAi10nBIJJfBOAC0x8hghBwbHVnvSKCEGRzdHK9sJJfBeAD+kAwIPpEAcjKB8v/ydDtRNCBAUDXIfQEMFyBAQj0Cm+hMbOSXwfgBdM/yCWCEHBsdWe6kjgw4w0DghBkc3RyupJfBuMNCgsCASAMDQBu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AAAK9ADJ7VQAeAH6APQEMPgnbyIwUAqhIb7y4FCCEHBsdWeDHrFwgBhQBMsFJs8WWPoCGfQAy2kXyx9SYMs/IMmAQPsABgCKUASBAQj0WTDtRNCBAUDXIMgBzxb0AMntVAFysI4jghBkc3Rygx6xcIAYUAXLBVADzxYj+gITy2rLH8s/yYBA+wCSXwPiAgEgDg8AWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWBARABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASASEwAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwGb/qfE=")
	require.NoError(t, err)
	codeCell, err := cell.FromBOC(codeBoC)
	require.NoError(t, err)

	i := &core.ContractInterface{
		Name:      known.NFTItem,
		Addresses: []*addr.Address{rndm.Address()},
		Code:      codeBoC,
		CodeHash:  codeCell.Hash(),
		GetMethodsDesc: []abi.GetMethodDesc{
			{
				Name: "get_nft_content",
				Arguments: []abi.VmValueDesc{
					{
						Name:      "index",
						StackType: "int",
					}, {
						Name:      "individual_content",
						StackType: "cell",
					},
				},
				ReturnValues: []abi.VmValueDesc{
					{
						Name:      "full_content",
						StackType: "cell",
						Format:    "content",
					},
				},
			},
		},
		GetMethodHashes: rndm.GetMethodHashes(),
	}

	schema := `{
        "op_name": "nft_item_transfer",
        "op_code": "0x5fcc3d14",
        "body": [
          {
            "name": "query_id",
            "tlb_type": "## 64"
          },
          {
            "name": "new_owner",
            "tlb_type": "addr"
          },
          {
            "name": "response_destination",
            "tlb_type": "addr"
          },
          {
            "name": "custom_payload",
            "tlb_type": "maybe ^"
          },
          {
            "name": "forward_amount",
            "tlb_type": ".",
            "format": "coins"
          },
          {
            "name": "forward_payload",
            "tlb_type": "either . ^",
            "format": "cell"
          }
        ]
      }`

	var opSchema abi.OperationDesc
	err = json.Unmarshal([]byte(schema), &opSchema)
	require.Nil(t, err)

	op := &core.ContractOperation{
		OperationName: "nft_item_transfer",
		ContractName:  known.NFTItem,
		MessageType:   core.Internal,
		Outgoing:      false,
		OperationID:   0xdeadbeed,
		Schema:        opSchema,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("insert definition", func(t *testing.T) {
		err := repo.AddDefinition(ctx, d.Name, d.Schema)
		require.Nil(t, err)
	})

	t.Run("get definitions", func(t *testing.T) {
		m, err := repo.GetDefinitions(ctx)
		require.Nil(t, err)
		require.Equal(t, 1, len(m))
		require.Equal(t, m[d.Name], d.Schema)
	})

	t.Run("insert interface", func(t *testing.T) {
		err := repo.AddInterface(ctx, i)
		require.Nil(t, err)
	})

	t.Run("insert operation", func(t *testing.T) {
		err := repo.AddOperation(ctx, op)
		require.Nil(t, err)
	})

	t.Run("get interfaces", func(t *testing.T) {
		ret, err := repo.GetInterfaces(ctx)
		require.Nil(t, err)
		require.Equal(t, []*core.ContractInterface{i}, ret)
	})

	t.Run("get operations", func(t *testing.T) {
		ret, err := repo.GetOperations(ctx)
		require.Nil(t, err)
		require.Equal(t, 1, len(ret))
		require.Equal(t, []*core.ContractOperation{op}, ret)
	})

	t.Run("get operation by id", func(t *testing.T) {
		ret, err := repo.GetOperationsByID(
			ctx,
			core.Internal,
			[]abi.ContractName{op.ContractName},
			op.Outgoing,
			op.OperationID,
		)
		require.Nil(t, err)
		require.Equal(t, op, ret[0])
	})

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})
}
