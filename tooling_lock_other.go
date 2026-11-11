//go:build !(unix || linux || darwin || freebsd || openbsd || netbsd || dragonfly)

package envai

import (
	"context"
	"errors"
)

func definitionFileLock(context.Context, string) (func(), error) {
	return nil, errors.New("envai: OS file locking unavailable; use PostgreSQL definition store on this platform")
}
