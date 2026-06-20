//go:build !windows || amd64 || 386 || arm

package db

import "github.com/uptrace/bun/driver/sqliteshim"

func driverName() string {
	return sqliteshim.ShimName
}
