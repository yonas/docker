// +build !exclude_graphdriver_zfs,linux,freebsd

package daemon

import (
	_ "github.com/docker/docker/daemon/graphdriver/zfs"
)
