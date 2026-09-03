package config

import (
	"strings"

	"github.com/gocql/gocql"
)

func ParseConsistencyLevel(name string) (gocql.Consistency, error) {
	if name == "" {
		return gocql.Quorum, nil
	}

	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "ANY":
		return gocql.Any, nil
	case "ONE":
		return gocql.One, nil
	case "TWO":
		return gocql.Two, nil
	case "THREE":
		return gocql.Three, nil
	case "QUORUM":
		return gocql.Quorum, nil
	case "LOCAL_ONE":
		return gocql.LocalOne, nil
	case "LOCAL_QUORUM":
		return gocql.LocalQuorum, nil
	case "EACH_QUORUM":
		return gocql.EachQuorum, nil
	case "ALL":
		return gocql.All, nil
	default:
		return gocql.Quorum, ErrInvalidConsistency
	}
}
