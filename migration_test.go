package migration

import (
	"context"
	"errors"
	"testing"

	"github.com/go-rel/rel"
	"github.com/go-rel/reltest"
	"github.com/stretchr/testify/assert"
)

func TestMigration(t *testing.T) {
	var (
		ctx       = context.TODO()
		repo      = reltest.New()
		migration = New(repo)
	)

	t.Run("Register", func(t *testing.T) {
		migration.Register(20200829084000,
			func(schema *rel.Schema) {
				schema.CreateTable("users", func(t *rel.Table) {
					t.ID("id")
				})
			},
			func(schema *rel.Schema) {
				schema.DropTable("users")
			},
		)

		migration.Register(20200828100000,
			func(schema *rel.Schema) {
				schema.CreateTable("tags", func(t *rel.Table) {
					t.ID("id")
				})

				schema.Do(func(repo rel.Repository) error {
					assert.NotNil(t, repo)
					return nil
				})
			},
			func(schema *rel.Schema) {
				schema.DropTable("tags")
			},
		)

		migration.Register(20200829115100,
			func(schema *rel.Schema) {
				schema.CreateTable("books", func(t *rel.Table) {
					t.ID("id")
				})
			},
			func(schema *rel.Schema) {
				schema.DropTable("books")
			},
		)

		assert.Len(t, migration.versions, 3)
		assert.Equal(t, 20200829084000, migration.versions[0].Version)
		assert.Equal(t, 20200828100000, migration.versions[1].Version)
		assert.Equal(t, 20200829115100, migration.versions[2].Version)

		repo.AssertExpectations(t)
	})

	t.Run("Migrate", func(t *testing.T) {
		repo.ExpectFindAll(rel.UsePrimary().SortAsc("version")).
			Result(versions{{ID: 1, Version: 20200829115100}})

		repo.ExpectTransaction(func(repo *reltest.Repository) {
			repo.ExpectInsert().For(&version{Version: 20200828100000})
		})

		repo.ExpectTransaction(func(repo *reltest.Repository) {
			repo.ExpectInsert().For(&version{Version: 20200829084000})
		})

		migration.Migrate(ctx)
		repo.AssertExpectations(t)
	})

	t.Run("Rollback", func(t *testing.T) {
		repo.ExpectFindAll(rel.UsePrimary().SortAsc("version")).
			Result(versions{
				{ID: 1, Version: 20200828100000},
				{ID: 2, Version: 20200829084000},
			})

		assert.Equal(t, 20200829084000, migration.versions[1].Version)

		repo.ExpectTransaction(func(repo *reltest.Repository) {
			repo.ExpectDelete().For(&migration.versions[1])
		})

		migration.Rollback(ctx)
		repo.AssertExpectations(t)
	})
}

func TestMigration_Sync(t *testing.T) {
	var (
		ctx  = context.TODO()
		repo = reltest.New()
		nfn  = func(schema *rel.Schema) {}
	)

	tests := []struct {
		name    string
		applied versions
		synced  versions
		isPanic bool
	}{
		{
			name: "all migrated",
			applied: versions{
				{ID: 1, Version: 1},
				{ID: 2, Version: 2},
				{ID: 3, Version: 3},
			},
			synced: versions{
				{ID: 1, Version: 1, applied: true},
				{ID: 2, Version: 2, applied: true},
				{ID: 3, Version: 3, applied: true},
			},
		},
		{
			name:    "not migrated",
			applied: versions{},
			synced: versions{
				{ID: 0, Version: 1, applied: false},
				{ID: 0, Version: 2, applied: false},
				{ID: 0, Version: 3, applied: false},
			},
		},
		{
			name: "first not migrated",
			applied: versions{
				{ID: 2, Version: 2},
				{ID: 3, Version: 3},
			},
			synced: versions{
				{ID: 0, Version: 1, applied: false},
				{ID: 2, Version: 2, applied: true},
				{ID: 3, Version: 3, applied: true},
			},
		},
		{
			name: "middle not migrated",
			applied: versions{
				{ID: 1, Version: 1},
				{ID: 3, Version: 3},
			},
			synced: versions{
				{ID: 1, Version: 1, applied: true},
				{ID: 0, Version: 2, applied: false},
				{ID: 3, Version: 3, applied: true},
			},
		},
		{
			name: "last not migrated",
			applied: versions{
				{ID: 1, Version: 1},
				{ID: 2, Version: 2},
			},
			synced: versions{
				{ID: 1, Version: 1, applied: true},
				{ID: 2, Version: 2, applied: true},
				{ID: 0, Version: 3, applied: false},
			},
		},
		{
			name: "broken migration",
			applied: versions{
				{ID: 1, Version: 1},
				{ID: 2, Version: 2},
				{ID: 3, Version: 3},
				{ID: 4, Version: 4},
			},
			synced: versions{
				{ID: 1, Version: 1, applied: true},
				{ID: 2, Version: 2, applied: true},
				{ID: 3, Version: 3, applied: true},
			},
			isPanic: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			migration := New(repo)
			migration.Register(3, nfn, nfn)
			migration.Register(2, nfn, nfn)
			migration.Register(1, nfn, nfn)

			repo.ExpectFindAll(rel.UsePrimary().SortAsc("version")).Result(test.applied)

			if test.isPanic {
				assert.Panics(t, func() {
					migration.sync(ctx)
				})
			} else {
				assert.NotPanics(t, func() {
					migration.sync(ctx)
				})

				assert.Equal(t, test.synced, migration.versions)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestMigration_Instrumentation(t *testing.T) {
	var (
		ctx  = context.TODO()
		repo = reltest.New()
		m    = New(repo)
	)

	m.Instrumentation(func(context.Context, string, string) func(error) { return nil })
	m.instrumenter.Observe(ctx, "test", "test")
}

func TestCheck(t *testing.T) {
	assert.Panics(t, func() {
		check(errors.New("error"))
	})
}
