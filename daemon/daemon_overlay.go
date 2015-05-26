// +build !exclude_graphdriver_overlay,!freebsd

package daemon

import (
	_ "github.com/docker/docker/daemon/graphdriver/overlay"
)
