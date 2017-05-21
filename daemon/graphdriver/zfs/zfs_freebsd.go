package zfs

import (
	"strings"
)

func getMountpoint(id string) string {
	maxlen := 12

	// we need to preserve filesystem suffix
	suffix := strings.SplitN(id, "-", 2)

	if len(suffix) > 1 {
		return id[:maxlen] + "-" + suffix[1]
	}

	return id[:maxlen]
}
