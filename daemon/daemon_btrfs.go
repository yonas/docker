// +build !exclude_graphdriver_btrfs,!freebsd

package daemon

import (
	_ "github.com/docker/docker/daemon/graphdriver/btrfs"
)
