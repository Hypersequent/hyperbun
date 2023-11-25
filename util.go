// https://github.com/uptrace/bun/blob/master/schema/table.go
// BSD 2-Clause License

package hyperbun

import (
	"reflect"
	"sync"

	"github.com/hypersequent/tms/packages/hyperbun/internal"
	"github.com/hypersequent/tms/packages/hyperbun/internal/tagparser"
	"github.com/uptrace/bun/schema"
)

var (
	baseModelType = reflect.TypeOf((*schema.BaseModel)(nil)).Elem()
	columnsCache  = sync.Map{}
)

func getColumns(typ reflect.Type) []string {
	if columns, ok := columnsCache.Load(typ); ok {
		return columns.([]string)
	}

	var columns []string

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		unexported := f.PkgPath != ""

		if unexported && !f.Anonymous { // unexported
			continue
		}
		if f.Tag.Get("bun") == "-" {
			continue
		}

		if f.Anonymous {
			if f.Name == "BaseModel" && f.Type == baseModelType {
				continue
			}

			fieldType := indirectType(f.Type)
			if fieldType.Kind() == reflect.Struct {
				// TODO: If field is an embedded struct, we should add each field of the embedded struct.
				continue
			}
		}

		// If field is not a struct, add it.
		// This will also add any embedded non-struct type as a field.
		columns = append(columns, fieldName(f))
	}

	columnsCache.Store(typ, columns)
	return columns
}

func fieldName(f reflect.StructField) string {
	tag := tagparser.Parse(f.Tag.Get("bun"))

	sqlName := internal.Underscore(f.Name)
	if tag.Name != "" && tag.Name != sqlName {
		sqlName = tag.Name
	}
	if s, ok := tag.Option("column"); ok {
		sqlName = s
	}

	return sqlName
}

func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
