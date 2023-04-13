package db

import (
	"context"
	"strings"
	"time"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/go-clickhouse/chmigrate"
	"github.com/urfave/cli/v2"

	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/migrations/ch"
	"github.com/tonindexer/anton/migrations/pg"
)

func newMigrators() (pg *migrate.Migrator, ch *chmigrate.Migrator, err error) {
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
			Name:  "create",
			Usage: "Creates up and down SQL migrations",
			Action: func(c *cli.Context) error {
				name := strings.Join(c.Args().Slice(), "_")

				mpg, mch, err := newMigrators()
				if err != nil {
					return err
				}

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
			Name:  "mark_applied",
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
	},
}
