package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/go-clickhouse/chmigrate"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/block"

	"github.com/tonindexer/anton/migrations/chmigrations"
	"github.com/tonindexer/anton/migrations/pgmigrations"
)

func newMigrators() (pg *migrate.Migrator, ck *chmigrate.Migrator, err error) {
	chURL := env.GetString("DB_CH_URL", "")
	pgURL := env.GetString("DB_PG_URL", "")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := repository.ConnectDB(ctx, chURL, pgURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot connect to the databases")
	}

	mpg := migrate.NewMigrator(conn.PG, pgmigrations.Migrations)
	mch := chmigrate.NewMigrator(conn.CH, chmigrations.Migrations)

	return mpg, mch, nil
}

func pgUnlock(ctx context.Context, m *migrate.Migrator) {
	if err := m.Unlock(ctx); err != nil {
		log.Error().Err(err).Msg("cannot unlock pg")
	}
}

func chUnlock(ctx context.Context, m *chmigrate.Migrator) {
	if err := m.Unlock(ctx); err != nil {
		log.Error().Err(err).Msg("cannot unlock ch")
	}
}

var Command = &cli.Command{
	Name:  "migrate",
	Usage: "Migrates database",

	Subcommands: []*cli.Command{
		{
			Name:  "init",
			Usage: "Creates migration tables",
			Action: func(c *cli.Context) error {
				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}
				if err := mpg.Init(c.Context); err != nil {
					return err
				}
				if err := mch.Init(c.Context); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:      "create",
			Usage:     "Creates up and down SQL migrations",
			ArgsUsage: "migration_name",
			Action: func(c *cli.Context) error {
				name := strings.Join(c.Args().Slice(), "_")

				mpg := migrate.NewMigrator(nil, pgmigrations.Migrations)
				mch := chmigrate.NewMigrator(nil, chmigrations.Migrations)

				pgFiles, err := mpg.CreateSQLMigrations(c.Context, name)
				if err != nil {
					return err
				}
				for _, mf := range pgFiles {
					log.Info().Str("name", mf.Name).Str("path", mf.Path).Msg("created pg migration")
				}

				chFiles, err := mch.CreateSQLMigrations(c.Context, name)
				if err != nil {
					return err
				}
				for _, mf := range chFiles {
					log.Info().Str("name", mf.Name).Str("path", mf.Path).Msg("created ch migration")
				}

				return nil
			},
		},
		{
			Name:  "up",
			Usage: "Migrates database",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "init",
					Aliases: []string{"i"},
					Value:   false,
					Usage:   "create migration tables",
				},
			},
			Action: func(c *cli.Context) error {
				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}

				if c.Bool("init") {
					if err := mpg.Init(c.Context); err != nil {
						return err
					}
					if err := mch.Init(c.Context); err != nil {
						return err
					}
				}

				// postgresql

				if err := mpg.Lock(c.Context); err != nil {
					return err
				}
				defer pgUnlock(c.Context, mpg)

				pgGroup, err := mpg.Migrate(c.Context)
				if err != nil {
					return err
				}
				if pgGroup.IsZero() {
					log.Info().Msg("there are no new migrations to run (pg database is up to date)")
				} else {
					log.Info().Str("group", pgGroup.String()).Msg("pg migrated")
				}

				// clickhouse

				if err := mch.Lock(c.Context); err != nil {
					return err
				}
				defer chUnlock(c.Context, mch)

				chGroup, err := mch.Migrate(c.Context)
				if err != nil {
					return err
				}
				if chGroup.IsZero() {
					log.Info().Msg("there are no new migrations to run (ch database is up to date)")
					return nil
				} else {
					log.Info().Str("group", chGroup.String()).Msg("ch migrated")
				}

				return nil
			},
		},
		{
			Name:  "down",
			Usage: "Rollbacks the last migration group",
			Action: func(c *cli.Context) error {
				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}

				// postgresql

				if err := mpg.Lock(c.Context); err != nil {
					return err
				}
				defer pgUnlock(c.Context, mpg)

				pgGroup, err := mpg.Rollback(c.Context)
				if err != nil {
					return err
				}
				if pgGroup.IsZero() {
					log.Info().Msg("there are no pg groups to roll back")
				} else {
					log.Info().Str("group", pgGroup.String()).Msg("pg rolled back")
				}

				// clickhouse

				if err := mch.Lock(c.Context); err != nil {
					return err
				}
				defer chUnlock(c.Context, mch)

				chGroup, err := mch.Rollback(c.Context)
				if err != nil {
					return err
				}
				if chGroup.IsZero() {
					log.Info().Msg("there are no ch groups to roll back")
				} else {
					log.Info().Str("group", chGroup.String()).Msg("ch rolled back")
				}

				return nil
			},
		},
		{
			Name:  "status",
			Usage: "Prints migrations status",
			Action: func(c *cli.Context) error {
				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}

				spg, err := mpg.MigrationsWithStatus(c.Context)
				if err != nil {
					return err
				}
				log.Info().Str("slice", spg.String()).Msg("pg all")
				log.Info().Str("slice", spg.Unapplied().String()).Msg("pg unapplied")
				log.Info().Str("group", spg.LastGroup().String()).Msg("pg last migration")

				sch, err := mch.MigrationsWithStatus(c.Context)
				if err != nil {
					return err
				}
				log.Info().Str("slice", sch.String()).Msg("ch all")
				log.Info().Str("slice", sch.Unapplied().String()).Msg("ch unapplied")
				log.Info().Str("group", sch.LastGroup().String()).Msg("ch last migration")

				return nil
			},
		},
		{
			Name:  "lock",
			Usage: "Locks migrations",
			Action: func(c *cli.Context) error {
				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}
				if err := mpg.Lock(c.Context); err != nil {
					return err
				}
				if err := mch.Lock(c.Context); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "unlock",
			Usage: "Unlocks migrations",
			Action: func(c *cli.Context) error {
				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}
				if err := mpg.Unlock(c.Context); err != nil {
					return err
				}
				if err := mch.Unlock(c.Context); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "markApplied",
			Usage: "Marks migrations as applied without actually running them",
			Action: func(c *cli.Context) error {
				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}
				pgGroup, err := mpg.Migrate(c.Context, migrate.WithNopMigration())
				if err != nil {
					return err
				}
				if pgGroup.IsZero() {
					log.Info().Msg("there are no new pg migrations to mark as applied")
					return nil
				}
				log.Info().Str("group", pgGroup.String()).Msg("pg marked as applied")

				chGroup, err := mch.Migrate(c.Context, chmigrate.WithNopMigration())
				if err != nil {
					return err
				}
				if chGroup.IsZero() {
					log.Info().Msg("there are no new ch migrations to mark as applied\n")
					return nil
				}
				log.Info().Str("group", pgGroup.String()).Msg("ch marked as applied")

				return nil
			},
		},
		{
			Name:  "transferCodeData",
			Usage: "Moves account states code and data from ClickHouse to RocksDB",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "limit",
					Value: 10000,
					Usage: "batch size for update",
				},
				&cli.Uint64Flag{
					Name:  "start-from",
					Value: 0,
					Usage: "last tx lt to start from",
				},
			},
			Action: func(c *cli.Context) error {
				chURL := env.GetString("DB_CH_URL", "")
				pgURL := env.GetString("DB_PG_URL", "")

				conn, err := repository.ConnectDB(c.Context, chURL, pgURL)
				if err != nil {
					return errors.Wrap(err, "cannot connect to the databases")
				}
				defer conn.Close()

				lastTxLT := c.Uint64("start-from")
				limit := c.Int("limit")
				for {
					var nextTxLT uint64

					err := conn.CH.NewRaw(`
							SELECT last_tx_lt
							FROM account_states
							WHERE last_tx_lt >= ?
							ORDER BY last_tx_lt ASC
							OFFSET ? ROW FETCH FIRST 1 ROWS ONLY`, lastTxLT, limit).
						Scan(c.Context, &nextTxLT)
					if errors.Is(err, sql.ErrNoRows) {
						nextTxLT = 0
						log.Info().Uint64("last_tx_lt", lastTxLT).Msg("finishing")
					} else if err != nil {
						return errors.Wrap(err, "get next tx lt")
					}

					f := fmt.Sprintf("last_tx_lt >= %d", lastTxLT)
					if nextTxLT != 0 {
						f += fmt.Sprintf(" AND last_tx_lt < %d", nextTxLT)
					}

					err = conn.CH.NewRaw(`
							INSERT INTO account_states_code
							SELECT code_hash, any(code)
							FROM (
								SELECT code_hash, code
								FROM account_states
								WHERE ` + f + ` AND length(code) > 0
							)
							GROUP BY code_hash`).
						Scan(c.Context)
					if err != nil {
						return errors.Wrapf(err, "transfer code from %d to %d", lastTxLT, nextTxLT)
					}

					err = conn.CH.NewRaw(`
							INSERT INTO account_states_data
							SELECT data_hash, any(data)
							FROM (
								SELECT data_hash, data
								FROM account_states
								WHERE ` + f + ` AND length(data) > 0
							)
							GROUP BY data_hash`).
						Scan(c.Context)
					if err != nil {
						return errors.Wrapf(err, "transfer data from %d to %d", lastTxLT, nextTxLT)
					}

					if nextTxLT == 0 {
						log.Info().Msg("finished")
						return nil
					} else {
						log.Info().Uint64("last_tx_lt", lastTxLT).Uint64("next_tx_lt", nextTxLT).Msg("transferred new batch")
					}
					lastTxLT = nextTxLT
				}
			},
		},
		{
			Name:  "clearCodeData",
			Usage: "Removes account states code and data from PostgreSQL",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "limit",
					Value: 10000,
					Usage: "batch size for update",
				},
				&cli.Uint64Flag{
					Name:  "start-from",
					Value: 0,
					Usage: "last tx lt to start from",
				},
			},
			Action: func(c *cli.Context) error {
				chURL := env.GetString("DB_CH_URL", "")
				pgURL := env.GetString("DB_PG_URL", "")

				conn, err := repository.ConnectDB(c.Context, chURL, pgURL)
				if err != nil {
					return errors.Wrap(err, "cannot connect to the databases")
				}
				defer conn.Close()

				lastTxLT := c.Uint64("start-from")
				limit := c.Int("limit")
				for {
					var nextTxLT uint64

					err := conn.CH.NewRaw(`
							SELECT last_tx_lt
							FROM account_states
							WHERE last_tx_lt >= ?
							ORDER BY last_tx_lt ASC
							OFFSET ? ROW FETCH FIRST 1 ROWS ONLY`, lastTxLT, limit).
						Scan(c.Context, &nextTxLT)
					if errors.Is(err, sql.ErrNoRows) {
						nextTxLT = 0
						log.Info().Uint64("last_tx_lt", lastTxLT).Msg("finishing")
					} else if err != nil {
						return errors.Wrap(err, "get next tx lt")
					}

					q := conn.PG.NewUpdate().Model((*core.AccountState)(nil)).
						Set("code = NULL").
						Where("last_tx_lt >= ?", lastTxLT)
					if nextTxLT != 0 {
						q = q.Where("last_tx_lt < ?", nextTxLT)
					}
					if _, err := q.Exec(c.Context); err != nil {
						return errors.Wrapf(err, "clear code from %d to %d", lastTxLT, nextTxLT)
					}

					q = conn.PG.NewUpdate().Model((*core.AccountState)(nil)).
						Set("data = NULL").
						Where("last_tx_lt >= ?", lastTxLT)
					if nextTxLT != 0 {
						q = q.Where("last_tx_lt < ?", nextTxLT)
					}
					if _, err := q.Exec(c.Context); err != nil {
						return errors.Wrapf(err, "clear data from %d to %d", lastTxLT, nextTxLT)
					}

					log.Info().Uint64("last_tx_lt", lastTxLT).Uint64("next_tx_lt", nextTxLT).Msg("cleared new batch")

					if nextTxLT == 0 {
						log.Info().Msg("finished")
						return nil
					}
					lastTxLT = nextTxLT
				}
			},
		},
		{
			Name:  "fillMissedClickHouseData",
			Usage: "Transfers missed blocks from PostgreSQL to ClickHouse",
			Action: func(c *cli.Context) error {
				chURL := env.GetString("DB_CH_URL", "")
				pgURL := env.GetString("DB_PG_URL", "")

				conn, err := repository.ConnectDB(c.Context, chURL, pgURL)
				if err != nil {
					return errors.Wrap(err, "cannot connect to the databases")
				}
				defer conn.Close()

				blockRepo := block.NewRepository(conn.CH, conn.PG)

				blockIds, err := blockRepo.GetMissedMasterBlocks(c.Context)
				if err != nil {
					return errors.Wrap(err, "get missed masterchain blocks")
				}
				if len(blockIds) == 0 {
					return errors.Wrap(core.ErrNotFound, "could not find any missed blocks")
				}

				m, err := blockRepo.GetLastMasterBlock(c.Context)
				if err != nil {
					return errors.Wrap(err, "get last master block")
				}

				for i := range blockIds {
					res, err := blockRepo.FilterBlocks(c.Context, &filter.BlocksReq{
						BlocksFilter: filter.BlocksFilter{
							Workchain: &m.Workchain,
							Shard:     &m.Shard,
							SeqNo:     &blockIds[i],
						},
						WithShards:              true,
						WithAccountStates:       true,
						WithTransactions:        true,
						WithTransactionMessages: true,
					})
					if err != nil {
						return errors.Wrapf(err, "filter blocks on (%d, %x, %d)", m.Workchain, m.Shard, m.SeqNo)
					}

					var (
						masterBlocks []*core.Block
						shardBlocks  []*core.Block
						transactions []*core.Transaction
						messages     []*core.Message
						accounts     []*core.AccountState
					)

					for _, row := range res.Rows {
						masterBlocks = append(masterBlocks, row)

						shardBlocks = append(shardBlocks, row.Shards...)

						accounts = append(accounts, row.Accounts...)
						transactions = append(transactions, row.Transactions...)
						for _, tx := range row.Transactions {
							if tx.InMsg != nil {
								messages = append(messages, tx.InMsg)
							}
							messages = append(messages, tx.OutMsg...)
						}

						for _, shard := range row.Shards {
							accounts = append(accounts, shard.Accounts...)
							transactions = append(transactions, shard.Transactions...)
							for _, tx := range shard.Transactions {
								transactions = append(transactions, tx)
								if tx.InMsg != nil {
									messages = append(messages, tx.InMsg)
								}
								messages = append(messages, tx.OutMsg...)
							}
						}
					}

					log.Info().
						Int32("workchain", m.Workchain).
						Int64("shard", m.Shard).
						Uint32("seq_no", blockIds[i]).
						Int("master_blocks_len", len(masterBlocks)).
						Int("shard_blocks_len", len(shardBlocks)).
						Int("transactions_len", len(transactions)).
						Int("messages_len", len(messages)).
						Int("account_states_len", len(accounts)).
						Msg("insert new missed block")

					if len(accounts) > 0 {
						if _, err := conn.CH.NewInsert().Model(&accounts).Exec(c.Context); err != nil {
							return err
						}
					}
					if len(messages) > 0 {
						if _, err := conn.CH.NewInsert().Model(&messages).Exec(c.Context); err != nil {
							return err
						}
					}
					if len(transactions) > 0 {
						if _, err := conn.CH.NewInsert().Model(&transactions).Exec(c.Context); err != nil {
							return err
						}
					}
					if len(shardBlocks) > 0 {
						if _, err := conn.CH.NewInsert().Model(&shardBlocks).Exec(c.Context); err != nil {
							return err
						}
					}
					if len(masterBlocks) > 0 {
						if _, err := conn.CH.NewInsert().Model(&masterBlocks).Exec(c.Context); err != nil {
							return err
						}
					}
				}

				return nil
			},
		},
		{
			Name:  "fillMissedRawAccountStates",
			Usage: "Fetches missed raw account states code and data",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "limit",
					Value: 10000,
					Usage: "batch size for update",
				},
				&cli.Uint64Flag{
					Name:  "start-from",
					Value: 0,
					Usage: "last tx lt to start from",
				},
			},
			Action: func(ctx *cli.Context) error {
				chURL := env.GetString("DB_CH_URL", "")
				pgURL := env.GetString("DB_PG_URL", "")

				conn, err := repository.ConnectDB(ctx.Context, chURL, pgURL)
				if err != nil {
					return errors.Wrap(err, "cannot connect to the databases")
				}
				defer conn.Close()

				client := liteclient.NewConnectionPool()
				api := ton.NewAPIClient(client, ton.ProofCheckPolicyUnsafe).WithRetry()
				for _, addr := range strings.Split(env.GetString("LITESERVERS", ""), ",") {
					split := strings.Split(addr, "|")
					if len(split) != 2 {
						return fmt.Errorf("wrong server address format '%s'", addr)
					}
					host, key := split[0], split[1]
					if err := client.AddConnection(ctx.Context, host, key); err != nil {
						return errors.Wrapf(err, "cannot add connection with %s host and %s key", host, key)
					}
				}

				fetchAccountCodeData := func(s *core.AccountState) error {
					block := core.Block{Workchain: s.Workchain, Shard: s.Shard, SeqNo: s.BlockSeqNo}

					err := conn.PG.NewSelect().Model(&block).
						Where("workchain = ?workchain").
						Where("shard = ?shard").
						Where("seq_no = ?seq_no").
						Scan(ctx.Context)
					if err != nil {
						return errors.Wrap(err, "select block")
					}

					rawBlock := ton.BlockIDExt{
						Workchain: block.Workchain, Shard: block.Shard, SeqNo: block.SeqNo,
						RootHash: block.RootHash, FileHash: block.FileHash,
					}

					acc, err := api.GetAccount(ctx.Context, &rawBlock, s.Address.MustToTonutils())
					if err != nil {
						return errors.Wrap(err, "get raw account")
					}

					if acc.Code == nil {
						return fmt.Errorf("account state has no code")
					}
					if acc.Data == nil {
						return fmt.Errorf("account state has no data")
					}

					code := core.AccountStateCode{
						CodeHash: acc.Code.Hash(),
						Code:     acc.Code.ToBOC(),
					}
					if _, err := conn.CH.NewInsert().Model(&code).Exec(ctx.Context); err != nil {
						return errors.Wrapf(err, "write code to key-value store")
					}

					data := core.AccountStateData{
						DataHash: acc.Data.Hash(),
						Data:     acc.Data.ToBOC(),
					}
					if _, err := conn.CH.NewInsert().Model(&data).Exec(ctx.Context); err != nil {
						return errors.Wrapf(err, "write data to key-value store")
					}

					return nil
				}

				getCodeData := func(ctx context.Context, rows []*core.AccountState) error {
					codeHashesSet, dataHashesSet := map[string]struct{}{}, map[string]struct{}{}
					for _, row := range rows {
						if len(row.CodeHash) == 32 {
							codeHashesSet[string(row.CodeHash)] = struct{}{}
						}
						if len(row.DataHash) == 32 {
							dataHashesSet[string(row.DataHash)] = struct{}{}
						}
					}

					batchLen := 1000
					codeHashBatches, dataHashBatches := make([][][]byte, 1), make([][][]byte, 1)
					appendHash := func(hash []byte, batches [][][]byte) [][][]byte {
						b := batches[len(batches)-1]
						if len(b) >= batchLen {
							b = [][]byte{}
							batches = append(batches, b)
						}
						batches[len(batches)-1] = append(b, hash)
						return batches
					}
					for h := range codeHashesSet {
						codeHashBatches = appendHash([]byte(h), codeHashBatches)
					}
					for h := range dataHashesSet {
						dataHashBatches = appendHash([]byte(h), dataHashBatches)
					}

					codeRes, dataRes := map[string][]byte{}, map[string][]byte{}
					for _, b := range codeHashBatches {
						var codeArr []*core.AccountStateCode
						err := conn.CH.NewSelect().Model(&codeArr).Where("code_hash IN ?", ch.In(b)).Scan(ctx)
						if err != nil {
							return errors.Wrapf(err, "get code")
						}
						for _, x := range codeArr {
							codeRes[string(x.CodeHash)] = x.Code
						}
					}
					for _, b := range dataHashBatches {
						var dataArr []*core.AccountStateData
						err := conn.CH.NewSelect().Model(&dataArr).Where("data_hash IN ?", ch.In(b)).Scan(ctx)
						if err != nil {
							return errors.Wrapf(err, "get data")
						}
						for _, x := range dataArr {
							dataRes[string(x.DataHash)] = x.Data
						}
					}

					for _, row := range rows {
						var ok bool
						if len(row.CodeHash) == 32 {
							if row.Code, ok = codeRes[string(row.CodeHash)]; !ok {
								log.Warn().
									Str("address", row.Address.String()).
									Uint64("last_tx_lt", row.LastTxLT).
									Msg("missed account code")
								if err := fetchAccountCodeData(row); err != nil {
									return errors.Wrapf(err, "(%s, %d)", row.Address.String(), row.LastTxLT)
								}
								continue
							}
						}
						if len(row.DataHash) == 32 {
							if row.Data, ok = dataRes[string(row.DataHash)]; !ok {
								log.Warn().
									Str("address", row.Address.String()).
									Uint64("last_tx_lt", row.LastTxLT).
									Msg("missed account data")
								if err := fetchAccountCodeData(row); err != nil {
									return errors.Wrapf(err, "(%s, %d)", row.Address.String(), row.LastTxLT)
								}
								continue
							}
						}
					}

					return nil
				}

				latestLT := ctx.Uint64("start-from")
				batch := ctx.Int("limit")
				totalChecked := 0
				for {
					var states []*core.AccountState

					err := conn.PG.NewSelect().Model(&states).
						Where("last_tx_lt > ?", latestLT).
						Order("last_tx_lt ASC").
						Limit(batch).
						Scan(ctx.Context)
					if err != nil {
						return errors.Wrapf(err, "scan account states from %d", latestLT)
					}

					if err := getCodeData(ctx.Context, states); err != nil {
						log.Error().Err(err).Msg("fill code data")
						continue
					}

					if len(states) < batch {
						log.Info().Msg("no states left")
						return nil
					}

					for _, s := range states {
						if s.LastTxLT > latestLT {
							latestLT = s.LastTxLT
						}
					}
					latestLT--

					totalChecked += len(states)
					if totalChecked%500000 == 0 {
						log.Info().Int("total_checked", totalChecked).Uint64("last_tx_lt", latestLT).Msg("checkpoint")
					}
				}
			},
		},
	},
}
