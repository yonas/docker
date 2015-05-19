// +build !exclude_graphdriver_devicemapper, !freebsd

package daemon

import (
	_ "github.com/docker/docker/daemon/graphdriver/devmapper"
)
