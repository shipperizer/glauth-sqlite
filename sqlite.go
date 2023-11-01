package main

import (
	"context"
	"database/sql"

	"github.com/glauth/glauth/v2/pkg/handler"
	"github.com/glauth/glauth/v2/pkg/plugins"
	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel/trace"
)

type SqliteBackend struct {
	tracer trace.Tracer
}

func NewSQLiteHandler(opts ...handler.Option) handler.Handler {
	options := newOptions(opts...)

	backend := SqliteBackend{
		tracer: options.Tracer,
	}

	return plugins.NewDatabaseHandler(backend, opts...)
}

func (b SqliteBackend) GetDriverName() string {
	return "sqlite3"
}

func (b SqliteBackend) GetPrepareSymbol() string {
	return "?"
}

// Create db/schema if necessary
func (b SqliteBackend) CreateSchema(ctx context.Context, db *sql.DB) {
	ctx, span := b.tracer.Start(ctx, "SqliteBackend.CreateSchema")
	defer span.End()

	statement, _ := db.Prepare(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	uidnumber INTEGER NOT NULL,
	primarygroup INTEGER NOT NULL,
	othergroups TEXT DEFAULT '',
	givenname TEXT DEFAULT '',
	sn TEXT DEFAULT '',
	mail TEXT DEFAULT '',
	loginshell TYEXT DEFAULT '',
	homedirectory TEXT DEFAULT '',
	disabled SMALLINT  DEFAULT 0,
	passsha256 TEXT DEFAULT '',
	passbcrypt TEXT DEFAULT '',
	otpsecret TEXT DEFAULT '',
	yubikey TEXT DEFAULT '',
	sshkeys TEXT DEFAULT '',
	custattr TEXT DEFAULT '{}')
`)
	statement.ExecContext(ctx)
	statement, _ = db.Prepare("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_name on users(name)")
	statement.ExecContext(ctx)
	statement, _ = db.Prepare("CREATE TABLE IF NOT EXISTS ldapgroups (id INTEGER PRIMARY KEY, name TEXT NOT NULL, gidnumber INTEGER NOT NULL)")
	statement.ExecContext(ctx)
	statement, _ = db.Prepare("CREATE UNIQUE INDEX IF NOT EXISTS idx_group_name on ldapgroups(name)")
	statement.ExecContext(ctx)
	statement, _ = db.Prepare("CREATE TABLE IF NOT EXISTS includegroups (id INTEGER PRIMARY KEY, parentgroupid INTEGER NOT NULL, includegroupid INTEGER NOT NULL)")
	statement.ExecContext(ctx)
	statement, _ = db.Prepare("CREATE TABLE IF NOT EXISTS capabilities (id INTEGER PRIMARY KEY, userid INTEGER NOT NULL, action TEXT NOT NULL, object TEXT NOT NULL)")
	statement.ExecContext(ctx)
}

// Migrate schema if necessary
func (b SqliteBackend) MigrateSchema(ctx context.Context, db *sql.DB, checker func(context.Context, *sql.DB, string, string) bool) {
	ctx, span := b.tracer.Start(ctx, "SqliteBackend.MigrateSchema")
	defer span.End()

	if !checker(ctx, db, "users", "sshkeys") {
		statement, _ := db.Prepare("ALTER TABLE users ADD COLUMN sshkeys TEXT DEFAULT ''")
		statement.ExecContext(ctx)
	}
	if checker(ctx, db, "groups", "name") {
		statement, _ := db.Prepare("DROP TABLE ldapgroups")
		statement.ExecContext(ctx)
		statement, _ = db.Prepare("ALTER TABLE groups RENAME TO ldapgroups")
		statement.ExecContext(ctx)
	}
}

// newOptions initializes the available default options.
func newOptions(opts ...handler.Option) handler.Options {
	opt := handler.Options{}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

func main() {}
