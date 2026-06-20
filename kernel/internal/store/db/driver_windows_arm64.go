//go:build windows && arm64

package db

import _ "modernc.org/sqlite"

func driverName() string {
	return "sqlite"
}
