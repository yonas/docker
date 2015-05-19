// +build freebsd

package daemon

import (
  "errors"
)

// checkExecSupport returns an error if the exec driver does not support exec,
// or nil if it is supported.
func checkExecSupport(DriverName string) error {
	return errors.New("not supported on bsd");
}
