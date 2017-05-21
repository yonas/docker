// +build linux freebsd

package zfs

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/mount"
	"github.com/docker/docker/pkg/parsers"
	zfs "github.com/mistifyio/go-zfs"
	"github.com/opencontainers/runc/libcontainer/label"
)

type zfsOptions struct {
	fsName    string
	mountPath string
}

func init() {
	graphdriver.Register("zfs", Init)
}

// Logger returns a zfs logger implementation.
type Logger struct{}

// Log wraps log message from ZFS driver with a prefix '[zfs]'.
func (*Logger) Log(cmd []string) {
	logrus.Debugf("[zfs] %s", strings.Join(cmd, " "))
}

// Init returns a new ZFS driver.
// It takes base mount path and a array of options which are represented as key value pairs.
// Each option is in the for key=value. 'zfs.fsname' is expected to be a valid key in the options.
func Init(base string, opt []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	var err error

	if _, err := exec.LookPath("zfs"); err != nil {
		logrus.Debugf("[zfs] zfs command is not available: %v", err)
		return nil, graphdriver.ErrPrerequisites
	}

	file, err := os.OpenFile("/dev/zfs", os.O_RDWR, 600)
	if err != nil {
		logrus.Debugf("[zfs] cannot open /dev/zfs: %v", err)
		return nil, graphdriver.ErrPrerequisites
	}
	defer file.Close()

	options, err := parseOptions(opt)
	if err != nil {
		return nil, err
	}
	options.mountPath = base

	rootdir := path.Dir(base)

	if options.fsName == "" {
		err = checkRootdirFs(rootdir)
		if err != nil {
			return nil, err
		}
	}

	if options.fsName == "" {
		options.fsName, err = lookupZfsDataset(rootdir)
		if err != nil {
			return nil, err
		}
	}

	zfs.SetLogger(new(Logger))

	filesystems, err := zfs.Filesystems(options.fsName)
	if err != nil {
		return nil, fmt.Errorf("Cannot find root filesystem %s: %v", options.fsName, err)
	}

	filesystemsCache := make(map[string]bool, len(filesystems))
	mountedFs := make(map[string]int, len(filesystems))

	var rootDataset *zfs.Dataset
	for _, fs := range filesystems {
		if fs.Name == options.fsName {
			rootDataset = fs
		}
		filesystemsCache[fs.Name] = true
		mountedFs[fs.Name] = 0
	}

	if rootDataset == nil {
		return nil, fmt.Errorf("BUG: zfs get all -t filesystem -rHp '%s' should contain '%s'", options.fsName, options.fsName)
	}

	d := &Driver{
		dataset:          rootDataset,
		options:          options,
		filesystemsCache: filesystemsCache,
		mountedFs:        mountedFs,
		uidMaps:          uidMaps,
		gidMaps:          gidMaps,
	}
	return graphdriver.NewNaiveDiffDriver(d, uidMaps, gidMaps), nil
}

func parseOptions(opt []string) (zfsOptions, error) {
	var options zfsOptions
	options.fsName = ""
	for _, option := range opt {
		key, val, err := parsers.ParseKeyValueOpt(option)
		if err != nil {
			return options, err
		}
		key = strings.ToLower(key)
		switch key {
		case "zfs.fsname":
			options.fsName = val
		default:
			return options, fmt.Errorf("Unknown option %s", key)
		}
	}
	return options, nil
}

func checkRootdirFs(rootdir string) error {
	var buf syscall.Statfs_t
	if err := syscall.Statfs(rootdir, &buf); err != nil {
		return fmt.Errorf("Failed to access '%s': %s", rootdir, err)
	}

	// FIXME: on freebsd buf.Type returns '0xDE', figure out why
	// if graphdriver.FsMagic(buf.Type) != graphdriver.FsMagicZfs {
	// 	log.Debugf("[zfs] no zfs dataset found for rootdir '%s'", rootdir)
	// 	return graphdriver.ErrPrerequisites
	// }
	return nil
}

func lookupZfsDataset(rootdir string) (string, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(rootdir, &stat); err != nil {
		return "", fmt.Errorf("Failed to access '%s': %s", rootdir, err)
	}
	wantedDev := stat.Dev

	mounts, err := mount.GetMounts()
	if err != nil {
		return "", err
	}
	for _, m := range mounts {
		if err := syscall.Stat(m.Mountpoint, &stat); err != nil {
			logrus.Debugf("[zfs] failed to stat '%s' while scanning for zfs mount: %v", m.Mountpoint, err)
			continue // may fail on fuse file systems
		}

		if stat.Dev == wantedDev && m.Fstype == "zfs" {
			return m.Source, nil
		}
	}

	return "", fmt.Errorf("Failed to find zfs dataset mounted on '%s' in /proc/mounts", rootdir)
}

// Driver holds information about the driver, such as zfs dataset, options and cache.
type Driver struct {
	dataset          *zfs.Dataset
	options          zfsOptions
	sync.Mutex       // protects filesystem cache against concurrent access
	filesystemsCache map[string]bool
	mountedFs 			 map[string]int
	uidMaps          []idtools.IDMap
	gidMaps          []idtools.IDMap
}

func (d *Driver) String() string {
	return "zfs"
}

// Cleanup is used to implement graphdriver.ProtoDriver. There is no cleanup required for this driver.
func (d *Driver) Cleanup() error {
	return nil
}

// Status returns information about the ZFS filesystem. It returns a two dimensional array of information
// such as pool name, dataset name, disk usage, parent quota and compression used.
// Currently it return 'Zpool', 'Zpool Health', 'Parent Dataset', 'Space Used By Parent',
// 'Space Available', 'Parent Quota' and 'Compression'.
func (d *Driver) Status() [][2]string {
	parts := strings.Split(d.dataset.Name, "/")
	pool, err := zfs.GetZpool(parts[0])

	var poolName, poolHealth string
	if err == nil {
		poolName = pool.Name
		poolHealth = pool.Health
	} else {
		poolName = fmt.Sprintf("error while getting pool information %v", err)
		poolHealth = "not available"
	}

	quota := "no"
	if d.dataset.Quota != 0 {
		quota = strconv.FormatUint(d.dataset.Quota, 10)
	}

	return [][2]string{
		{"Zpool", poolName},
		{"Zpool Health", poolHealth},
		{"Parent Dataset", d.dataset.Name},
		{"Space Used By Parent", strconv.FormatUint(d.dataset.Used, 10)},
		{"Space Available", strconv.FormatUint(d.dataset.Avail, 10)},
		{"Parent Quota", quota},
		{"Compression", d.dataset.Compression},
	}
}

// GetMetadata returns image/container metadata related to graph driver
func (d *Driver) GetMetadata(id string) (map[string]string, error) {
	return nil, nil
}

func (d *Driver) cloneFilesystem(name, parentName string) error {
	snapshotName := fmt.Sprintf("%d", time.Now().Nanosecond())
	parentDataset := zfs.Dataset{Name: parentName}
	snapshot, err := parentDataset.Snapshot(snapshotName /*recursive */, false)
	if err != nil {
		return err
	}

	_, err = snapshot.Clone(name, map[string]string{"mountpoint": "legacy"})
	if err == nil {
		d.Lock()
		d.filesystemsCache[name] = true
		d.mountedFs[name] = 0
		d.Unlock()
	}

	if err != nil {
		snapshot.Destroy(zfs.DestroyDeferDeletion)
		return err
	}
	return snapshot.Destroy(zfs.DestroyDeferDeletion)
}

func (d *Driver) zfsPath(id string) string {
	return d.options.fsName + "/" + id
}

func (d *Driver) MountPath(id string) string {
	newid := id;

	// on freebsd mount path is limited with 88 chars, so we need to use short ids
  if(runtime.GOOS == "freebsd") {
		suffix := strings.SplitN(id, "-", 2)
	
		if(len(suffix) == 1) {// no tag
			newid = id[0:12]
		} else { 
			newid = id[0:12] + "-" + suffix[1]
		}
  }   

	return path.Join(d.options.mountPath, "graph", newid)
}

// Create prepares the dataset and filesystem for the ZFS driver for the given id under the parent.
func (d *Driver) Create(id string, parent string, mountLabel string) error {
	err := d.create(id, parent)
	if err == nil {
		return nil
	}
	if zfsError, ok := err.(*zfs.Error); ok {
		if !strings.HasSuffix(zfsError.Stderr, "dataset already exists\n") {
			return err
		}
		// aborted build -> cleanup
	} else {
		return err
	}

	dataset := zfs.Dataset{Name: d.zfsPath(id)}
	if err := dataset.Destroy(zfs.DestroyRecursiveClones); err != nil {
		return err
	}

	// retry
	return d.create(id, parent)
}

func (d *Driver) create(id, parent string) error {
	name := d.zfsPath(id)
	if parent == "" {
		mountoptions := map[string]string{"mountpoint": "legacy"}
		fs, err := zfs.CreateFilesystem(name, mountoptions)
		if err == nil {
			d.Lock()
			d.filesystemsCache[fs.Name] = true
			d.mountedFs[fs.Name] = 0
			d.Unlock()
		}
		return err
	}
	return d.cloneFilesystem(name, d.zfsPath(parent))
}

// Remove deletes the dataset, filesystem and the cache for the given id.
func (d *Driver) Remove(id string) error {
	name := d.zfsPath(id)
	dataset := zfs.Dataset{Name: name}
	err := dataset.Destroy(zfs.DestroyRecursive)
	if err == nil {
		d.Lock()
		delete(d.filesystemsCache, name)
		delete(d.mountedFs, name)
		d.Unlock()
	}
	return err
}

// Get returns the mountpoint for the given id after creating the target directories if necessary.
func (d *Driver) Get(id, mountLabel string) (string, error) {
	mountpoint := d.MountPath(id)
	filesystem := d.ZfsPath(id)

	if(d.mountedFs[filesystem] == 0) {
		log.Debugf(`[zfs] mount("%s", "%s", "%s")`, filesystem, mountpoint, mountLabel)

		// Create the target directories if they don't exist
		if err := os.MkdirAll(mountpoint, 0755); err != nil && !os.IsExist(err) {
			return "", err
		}

		err := mount.Mount(filesystem, mountpoint, "zfs", mountLabel)
		if err != nil {
			return "", fmt.Errorf("error creating zfs mount of %s to %s: %v", filesystem, mountpoint, err)
		}
	} else {
		log.Debugf("[zfs] using already mounted fs %s", id)
	}

	d.Lock()
	d.mountedFs[filesystem]++
	d.Unlock()

	return mountpoint, nil
}

// Put removes the existing mountpoint for the given id if it exists.
func (d *Driver) Put(id string) error {
	mountpoint := d.MountPath(id)
	name := d.ZfsPath(id)

	if(d.mountedFs[name] == 1) {
		log.Debugf(`[zfs] unmount("%s")`, mountpoint)

		if err := mount.Unmount(mountpoint); err != nil {
			return fmt.Errorf("error unmounting to %s: %v", mountpoint, err)
		}
	} else {
		log.Debugf("[zfs] fs still used: %s", id)
	}

	d.Lock()
	d.mountedFs[name]--
	d.Unlock()

	return nil
}

// Exists checks to see if the cache entry exists for the given id.
func (d *Driver) Exists(id string) bool {
	return d.filesystemsCache[d.zfsPath(id)] == true
}
