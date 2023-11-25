package hyperbun

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

type Entity struct {
	bun.BaseModel `bun:"table:entities"`

	ID   int `bun:",nullzero"`
	Name string
}

func TestGetColumns(t *testing.T) {
	assert.Equal(t, []string{"id", "name"}, getColumns(reflect.TypeOf(Entity{})))
}
