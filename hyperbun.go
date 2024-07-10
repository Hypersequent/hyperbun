package hyperbun

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"

	"github.com/uptrace/bun"

	"github.com/hypersequent/hyperr"
)

type DB interface {
	Context() context.Context
	NewSelect() *bun.SelectQuery
	NewInsert() *bun.InsertQuery
	NewUpdate() *bun.UpdateQuery
	NewDelete() *bun.DeleteQuery
	NewMerge() *bun.MergeQuery
	NewRaw(string, ...interface{}) *bun.RawQuery
	NewValues(model interface{}) *bun.ValuesQuery
	RunInTx(fn func(tx TxContext) error) error
}

type Context struct {
	ctx context.Context
	Bun *bun.DB
}

var _ DB = (*Context)(nil)

func NewContext(ctx context.Context, db *bun.DB) *Context {
	return &Context{
		ctx: ctx,
		Bun: db,
	}
}

func (m Context) Context() context.Context {
	return m.ctx
}

func (m Context) NewSelect() *bun.SelectQuery {
	return m.Bun.NewSelect()
}

func (m Context) NewInsert() *bun.InsertQuery {
	return m.Bun.NewInsert()
}

func (m Context) NewUpdate() *bun.UpdateQuery {
	return m.Bun.NewUpdate()
}

func (m Context) NewDelete() *bun.DeleteQuery {
	return m.Bun.NewDelete()
}

func (m Context) NewMerge() *bun.MergeQuery {
	return m.Bun.NewMerge()
}

func (m Context) NewRaw(query string, args ...interface{}) *bun.RawQuery {
	return m.Bun.NewRaw(query, args...)
}

func (m Context) NewValues(model interface{}) *bun.ValuesQuery {
	return m.Bun.NewValues(model)
}

func (m Context) RunInTx(fn func(tx TxContext) error) error {
	return m.Bun.RunInTx(m.ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		return fn(NewTxContext(ctx, tx))
	})
}

// ---------------------------------------------------------------

type TxContext struct {
	ctx context.Context
	Bun bun.Tx
}

var _ DB = (*TxContext)(nil)

func NewTxContext(ctx context.Context, tx bun.Tx) TxContext {
	return TxContext{
		ctx: ctx,
		Bun: tx,
	}
}

func (m TxContext) Context() context.Context {
	return m.ctx
}

func (m TxContext) NewSelect() *bun.SelectQuery {
	return m.Bun.NewSelect()
}

func (m TxContext) NewInsert() *bun.InsertQuery {
	return m.Bun.NewInsert()
}

func (m TxContext) NewUpdate() *bun.UpdateQuery {
	return m.Bun.NewUpdate()
}

func (m TxContext) NewDelete() *bun.DeleteQuery {
	return m.Bun.NewDelete()
}

func (m TxContext) NewMerge() *bun.MergeQuery {
	return m.Bun.NewMerge()
}

func (m TxContext) NewRaw(query string, args ...interface{}) *bun.RawQuery {
	return m.Bun.NewRaw(query, args...)
}

func (m TxContext) NewValues(model interface{}) *bun.ValuesQuery {
	return m.Bun.NewValues(model)
}

func (m TxContext) RunInTx(fn func(tx TxContext) error) error {
	return m.Bun.RunInTx(m.ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		return fn(NewTxContext(ctx, tx))
	})
}

func ByID[T any, ID string | int](m DB, id ID) (*T, error) {
	var row T
	if err := m.NewSelect().
		Model(&row).
		Where("id = ?", id).
		Limit(1).
		Scan(m.Context()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, annotate(err, "ByID", "table", hyperbunTableForType[T](), "id", id)
	}

	return &row, nil
}

func StructByID[T any, ID string | int](m DB, table string, id ID) (*T, error) {
	var row T
	columns := getColumns(reflect.TypeOf(row))
	if err := m.NewSelect().
		Column(columns...).
		Table(table).
		Where("id = ?", id).
		Limit(1).
		Scan(m.Context(), &row); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, annotate(err, "StructByID", "table", table, "id", id)
	}

	return &row, nil
}

func TypeByID[T any, ID string | int](m DB, table string, column string, id ID) (*T, error) {
	var value T
	if err := m.NewSelect().
		ColumnExpr(column).
		Table(table).
		Where("id = ?", id).
		Limit(1).
		Scan(m.Context(), &value); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, annotate(err, "TypeByID", "table", table, "column", column, "id", id)
	}

	return &value, nil
}

func BySQL[T any](m DB, query string, args ...interface{}) (*T, error) {
	var row T
	if err := m.NewSelect().
		Model(&row).
		Where(query, args...).
		Limit(1).
		Scan(m.Context()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, annotate(err, "BySQL", "table", hyperbunTableForType[T]())
	}

	return &row, nil
}

func StructBySQL[T any](m DB, table string, query string, args ...interface{}) (*T, error) {
	var row T
	columns := getColumns(reflect.TypeOf(row))
	if err := m.NewSelect().
		Column(columns...).
		Table(table).
		Where(query, args...).
		Limit(1).
		Scan(m.Context(), &row); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, annotate(err, "StructBySQL", "table", table)
	}

	return &row, nil
}

func TypeBySQL[T any](m DB, table string, column string, query string, args ...interface{}) (*T, error) {
	var value T
	if err := m.NewSelect().
		ColumnExpr(column).
		Table(table).
		Where(query, args...).
		Limit(1).
		Scan(m.Context(), &value); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, annotate(err, "TypeBySQL", "table", table, "column", column)
	}

	return &value, nil
}

func Many[T any](m DB, query string, args ...interface{}) ([]T, error) {
	var rows []T
	if err := m.NewSelect().
		Model(&rows).
		Where(query, args...).
		Scan(m.Context()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, annotate(err, "Many", "table", hyperbunTableForType[T]())
	}

	return rows, nil
}

func Exists[ID string | int](m DB, table string, id ID) (bool, error) {
	c, err := CountQuery(m, table, "id = ?", id)
	if err != nil {
		return false, annotate(err, "Exists", "table", table, "id", id)
	}

	return c == 1, nil
}

func ExistsBySQL(m DB, table string, query string, args ...interface{}) (bool, error) {
	var exists bool
	if err := m.NewRaw("SELECT EXISTS(SELECT 1 from "+table+" WHERE "+query+")", args...).
		Scan(m.Context(), &exists); err != nil {
		return false, annotate(err, "ExistsBySQL", "table", table)
	}

	return exists, nil
}

func CountQuery(m DB, table string, query string, args ...interface{}) (int, error) {
	count, err := m.NewSelect().
		Table(table).
		Where(query, args...).
		Count(m.Context())
	if err != nil {
		return 0, annotate(err, "CountQuery", "table", table)
	}
	return count, nil
}

func Insert[T any](m DB, row *T) error {
	_, err := m.NewInsert().
		Model(row).
		Exec(m.Context())
	if err != nil {
		return annotate(err, "Insert", "table", hyperbunTableForType[T]())
	}
	return nil
}

func InsertMany[T any](m DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}

	_, err := m.NewInsert().
		Model(&rows).
		Exec(m.Context())
	if err != nil {
		return annotate(err, "InsertMany", "table", hyperbunTableForType[T]())
	}
	return nil
}

func Update[T any](m DB, row *T, pk ...string) error {
	if len(pk) == 0 {
		pk = append(pk, "id")
	}
	if _, err := m.NewUpdate().
		Model(row).
		WherePK(pk...).
		Exec(m.Context()); err != nil {
		return annotate(err, "Update", "table", hyperbunTableForType[T](), "pk", strings.Join(pk, ","))
	}
	return nil
}

func UpdateSQLByID[ID string | int](m DB, table string, id ID, query string, args ...interface{}) error {
	_, err := m.NewUpdate().
		Table(table).
		Set(query, args...).
		Where("id = ?", id).
		Exec(m.Context())
	if err != nil {
		return annotate(err, "UpdateSQLByID", "table", table, "id", id)
	}
	return err
}

// To upsert and check multiple constraints, see
// https://stackoverflow.com/questions/35888012/use-multiple-conflict-target-in-on-conflict-clause
func Upsert[T any](m DB, rows T, conflictColumns string) error {
	if _, err := m.NewInsert().
		Model(rows).
		On(fmt.Sprintf("conflict (%s) do update", conflictColumns)).
		Exec(m.Context()); err != nil {
		return annotate(err, "Upsert", "table", hyperbunTableForType[T](), "conflict", conflictColumns)
	}
	return nil
}

func UpsertIgnore[T any](m DB, rows T) error {
	_, err := m.NewInsert().
		Model(rows).
		On("conflict do nothing").
		Exec(m.Context())
	if err != nil {
		return annotate(err, "UpsertIgnore", "table", hyperbunTableForType[T]())
	}

	return err
}

func DeleteByID[ID string | int](m DB, table string, id ID) error {
	if _, err := m.NewDelete().
		Table(table).
		Where("id = ?", id).
		Exec(m.Context()); err != nil {
		return annotate(err, "DeleteByID", "table", table, "id", id)
	}

	return nil
}

func DeleteBySQL(m DB, table string, query string, args ...interface{}) error {
	if _, err := m.NewDelete().
		Table(table).
		Where(query, args...).
		Exec(m.Context()); err != nil {
		return annotate(err, "DeleteBySQL", "table", table)
	}

	return nil
}

func RunInTx(m DB, fn func(tx TxContext) error) error {
	if err := m.RunInTx(fn); err != nil {
		return fmt.Errorf("RunInTx: %w", err)
	}
	return nil
}

func RunInLockedTx(m DB, id string, fn func(tx TxContext) error) error {
	return RunInTx(m, func(tx TxContext) error {
		if err := advisoryLock(m, id); err != nil {
			return hyperr.Wrap(err)
		}

		return fn(tx)
	})
}

func advisoryLock(m DB, name string) error {
	h := fnv.New64()
	h.Write([]byte(name))
	s := h.Sum64()
	if _, err := m.NewRaw("SELECT pg_advisory_xact_lock(?)", int64(s)).
		Exec(m.Context()); err != nil {
		return hyperr.Wrap(err)
	}

	return nil
}

func annotate(err error, op string, kvs ...interface{}) error {
	pairs := make([][2]string, len(kvs)/2)
	numPairs := len(kvs) / 2
	odd := len(kvs)%2 != 0
	for i := 0; i < numPairs; i++ {
		pairs[i] = [2]string{
			fmt.Sprint(kvs[i*2]),
			fmt.Sprint(kvs[i*2+1]),
		}
	}
	if odd {
		pairs = append(pairs, [2]string{
			fmt.Sprint(kvs[len(kvs)-1]),
			"<missing value>",
		})
	}
	joined := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		joined = append(joined, fmt.Sprint(pair[0], "='", pair[1], "'"))
	}
	joinedStr := strings.Join(joined, " ")
	if joinedStr != "" {
		joinedStr = " " + joinedStr
	}

	return fmt.Errorf("performing %s%s: %w", op, joinedStr, err)
}

func hyperbunTableForType[T any]() string {
	var t T
	typ := reflect.TypeOf(t)
	kind := typ.Kind()

	// This covers the case like *U, []U, *[]U, []*U etc
	for kind == reflect.Pointer || kind == reflect.Slice {
		typ = typ.Elem()
		kind = typ.Kind()
	}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		val, ok := f.Tag.Lookup("bun")
		if !ok {
			continue
		}
		for _, ann := range strings.Split(val, ",") {
			spl := strings.Split(ann, ":")
			if len(spl) != 2 {
				continue
			}
			if spl[0] == "table" {
				return spl[1]
			}
		}
	}
	return ""
}
