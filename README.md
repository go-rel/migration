# migration

[![GoDoc](https://godoc.org/github.com/go-rel/migration?status.svg)](https://pkg.go.dev/github.com/go-rel/migration)
[![Test](https://github.com/go-rel/migration/actions/workflows/test.yml/badge.svg)](https://github.com/go-rel/migration/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-rel/migration)](https://goreportcard.com/report/github.com/go-rel/migration)
[![codecov](https://codecov.io/gh/go-rel/migration/branch/main/graph/badge.svg?token=1yyLz5sbBR)](https://codecov.io/gh/go-rel/migration)
[![Gitter chat](https://badges.gitter.im/go-rel/rel.png)](https://gitter.im/go-rel/rel)

Database Migration utility for Golang.

## Example 

```go
package main

import (
    "context"

    "github.com/go-rel/doc/examples/db/migrations"
    "github.com/go-rel/mysql"
    "github.com/go-rel/rel"
    "github.com/go-rel/rel/migrator"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    var (
        ctx  = context.TODO()
        repo = rel.New(mysql.MustOpen("root@(source:3306)/rel_test?charset=utf8&parseTime=True&loc=Local"))
        m    = migrator.New(repo)
    )

    // Register migrations
    m.Register(20202806225100, migrations.MigrateCreateTodos, migrations.RollbackCreateTodos)

    // Run migrations
    m.Migrate(ctx)
    // OR:
    // m.Rollback(ctx)
}
```

More Info: https://go-rel.github.io/migration

## License

Released under the [MIT License](https://github.com/go-rel/migration/blob/master/LICENSE)
