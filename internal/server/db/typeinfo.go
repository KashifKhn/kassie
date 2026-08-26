package db

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
)

func CQLTypeString(ti gocql.TypeInfo) string {
	switch t := ti.(type) {
	case gocql.CollectionType:
		switch t.Type() {
		case gocql.TypeMap:
			return fmt.Sprintf("map<%s, %s>", CQLTypeString(t.Key), CQLTypeString(t.Elem))
		case gocql.TypeList:
			return fmt.Sprintf("list<%s>", CQLTypeString(t.Elem))
		case gocql.TypeSet:
			return fmt.Sprintf("set<%s>", CQLTypeString(t.Elem))
		default:
			return t.String()
		}
	case gocql.TupleTypeInfo:
		parts := make([]string, len(t.Elems))
		for i, elem := range t.Elems {
			parts[i] = CQLTypeString(elem)
		}
		return fmt.Sprintf("tuple<%s>", strings.Join(parts, ", "))
	default:
		return ti.Type().String()
	}
}
