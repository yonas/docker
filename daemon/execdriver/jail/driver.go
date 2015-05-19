package jail

import (
	//"fmt"
	"io/ioutil"
	//"log"
	"os"
	"os/exec"
	"path"
	//"runtime"
	"syscall"

	"github.com/docker/docker/daemon/execdriver"

	"github.com/Sirupsen/logrus"
)

const DriverName = "jail"
const Version = "0.1"

func init() {
	// execdriver.RegisterInitFunc(DriverName, func(args *execdriver.InitArgs) error {
	// 	runtime.LockOSThread()

	// 	path, err := exec.LookPath(args.Args[0])
	// 	if err != nil {
	// 		log.Printf("Unable to locate %v", args.Args[0])
	// 		os.Exit(127)
	// 	}
	// 	if err := syscall.Exec(path, args.Args, os.Environ()); err != nil {
	// 		return fmt.Errorf("dockerinit unable to execute %s - %s", path, err)
	// 	}
	// 	panic("Unreachable")
	// })
}

type driver struct {
	root     string
	initPath string
}

func NewDriver(root, initPath string) (*driver, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}

	return &driver{
		root:     root,
		initPath: initPath,
	}, nil
}

func (d *driver) Name() string {
	return DriverName
}

func copyFile(src string, dest string) error {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, content, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) Run(c *execdriver.Command, pipes *execdriver.Pipes, startCallback execdriver.StartCallback) (execdriver.ExitStatus, error) {
	// if err := execdriver.SetTerminal(c, pipes); err != nil {
	// 	return -1, err
	// }

	logrus.Info("running jail")

	root := c.Rootfs

	init := path.Join(root, ".dockerinit")
	if err := copyFile(os.Args[0], init); err != nil {
		return execdriver.ExitStatus{-1, false}, err
	}

	devDir := path.Join(root, "dev")
	if err := os.MkdirAll(devDir, 0755); err != nil {
		return execdriver.ExitStatus{-1, false}, err
	}

	params := []string{
		"/usr/sbin/jail",
		"-c",
		"name=" + c.ID,
		"path=" + root,
		//"command=" + c.InitPath,
	}

	// if c.ProcessConfig.User != "" {
	// 	params = append(params, "-u", c.ProcessConfig.User)
	// }

	// if c.ProcessConfig.Privileged {
	// 	params = append(params, "-privileged")
	// }

	// if c.WorkingDir != "" {
	// 	params = append(params, "-w", c.WorkingDir)
	// }

	params = append(params, "command=csh")//, c.ProcessConfig.Entrypoint)
	params = append(params, c.ProcessConfig.Arguments...)

	c.ProcessConfig.Path = "/usr/sbin/jail"
	c.ProcessConfig.Args = params

	logrus.Debugf("jail params %s", params)

	if err := c.ProcessConfig.Run(); err != nil {
		logrus.Infof("jail failed %s", err)
		return execdriver.ExitStatus{-1, false}, err
	}

	return execdriver.ExitStatus{getExitCode(c), false}, nil
}

func getExitCode(c *execdriver.Command) int {
	if c.ProcessConfig.ProcessState == nil {
		return -1
	}
	return c.ProcessConfig.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
}

func (d *driver) Kill(c *execdriver.Command, sig int) error {
	return nil
}

func (d *driver) Pause(c *execdriver.Command) error {
	return nil
}

func (d *driver) Unpause(c *execdriver.Command) error {
	return nil
}

func (d *driver) Terminate(c *execdriver.Command) error {
	return nil
}

func (d *driver) GetPidsForContainer(id string) ([]int, error) {
	return nil, nil
}

func (d *driver) Clean(id string) error {
	return nil
}

func (d *driver) Exec(c *execdriver.Command, processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes, startCallback execdriver.StartCallback) (int, error) {
	return 0, nil
}

func (d *driver) Stats(id string) (*execdriver.ResourceStats, error)  {
	return nil, nil
}

type info struct {
	ID     string
	driver *driver
}

func (d *driver) Info(id string) execdriver.Info {
	return &info{ID: id, driver: d}
}

func (info *info) IsRunning() bool {
	if err := exec.Command("jls", "-j", info.ID).Run(); err != nil {
		return true
	}

	return false
}
