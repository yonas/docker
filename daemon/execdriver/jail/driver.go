package jail

import (
	//"fmt"
	"io/ioutil"
	//"log"
	"os"
	"os/exec"
	//"path"
	//"runtime"
	"syscall"

	"github.com/docker/docker/daemon/execdriver"

	"github.com/Sirupsen/logrus"


	"github.com/kr/pty"
	"io"
	"github.com/docker/docker/pkg/term"
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
	var (
		term    execdriver.Terminal
		err 		error
	)

	// setting terminal parameters
	if c.ProcessConfig.Tty {
		term, err = NewTtyConsole(&c.ProcessConfig, pipes)
	} else {
		term, err = execdriver.NewStdConsole(&c.ProcessConfig, pipes)
	}
	if err != nil {
		return execdriver.ExitStatus{ExitCode: -1}, err
	}
	c.ProcessConfig.Terminal = term

	logrus.Info("running jail")

	

	// init := path.Join(root, ".dockerinit")
	// if err := copyFile(os.Args[0], init); err != nil {
	// 	return execdriver.ExitStatus{ExitCode: -1}, err
	// }

	// // create /dev
	root := c.Rootfs
	// devDir := path.Join(root, "dev")
	// if err := os.MkdirAll(devDir, 0755); err != nil {
	// 	return execdriver.ExitStatus{ExitCode: -1}, err
	// }

	// build params for the jail
	params := []string{
		"/usr/sbin/jail",
		"-c",
		"name=" + c.ID,
		"path=" + root,
		"mount.devfs=1",
		"command=" + c.ProcessConfig.Entrypoint,
	}

	params = append(params, c.ProcessConfig.Arguments...)

	c.ProcessConfig.Path = "/usr/sbin/jail"
	c.ProcessConfig.Args = params

	logrus.Debugf("jail params %s", params)

	if err := c.ProcessConfig.Start(); err != nil {
		logrus.Infof("jail failed %s", err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	logrus.Debug("jail started");

  //=====


	var (
		waitErr  error
		waitLock = make(chan struct{})
	)

	go func() {
		if err := c.ProcessConfig.Wait(); err != nil {
			if _, ok := err.(*exec.ExitError); !ok { // Do not propagate the error if it's simply a status code != 0
				waitErr = err
			}
		}
		close(waitLock)
	}()

	var pid int

	// terminate := func(terr error) (execdriver.ExitStatus, error) {
	// 	if c.ProcessConfig.Process != nil {
	// 		c.ProcessConfig.Process.Kill()
	// 		c.ProcessConfig.Wait()
	// 	}
	// 	return execdriver.ExitStatus{ExitCode: -1}, terr
	// }
	// // Poll lxc for RUNNING status
	// pid, err := d.waitForStart(c, waitLock)
	// if err != nil {
	// 	return terminate(err)
	// }

	// cgroupPaths, err := cgroupPaths(c.ID)
	// if err != nil {
	// 	return terminate(err)
	// }

	// state := &libcontainer.State{
	// 	InitProcessPid: pid,
	// 	CgroupPaths:    cgroupPaths,
	// }

	// f, err := os.Create(filepath.Join(dataPath, "state.json"))
	// if err != nil {
	// 	return terminate(err)
	// }
	// defer f.Close()

	// if err := json.NewEncoder(f).Encode(state); err != nil {
	// 	return terminate(err)
	// }

	c.ContainerPid = pid

	if startCallback != nil {
		logrus.Debugf("Invoking startCallback")
		startCallback(&c.ProcessConfig, pid)
	}

	// oomKill := false
	// oomKillNotification, err := notifyOnOOM(cgroupPaths)

	<-waitLock
	exitCode := getExitCode(c)

	if err := exec.Command("umount", root + "/dev").Run(); err != nil { 		
		logrus.Debugf("umount %s failed: %s", c.ID, err);
	}


	// if err == nil {
	// 	_, oomKill = <-oomKillNotification
	// 	logrus.Debugf("oomKill error: %v, waitErr: %v", oomKill, waitErr)
	// } else {
	// 	logrus.Warnf("Your kernel does not support OOM notifications: %s", err)
	// }

	// // check oom error
	// if oomKill {
	// 	exitCode = 137
	// }

	return execdriver.ExitStatus{ExitCode: exitCode, OOMKilled: false}, waitErr

  //=====

	//return execdriver.ExitStatus{getExitCode(c), false}, nil
}

func getExitCode(c *execdriver.Command) int {
	if c.ProcessConfig.ProcessState == nil {
		return -1
	}
	return c.ProcessConfig.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
}

func (d *driver) Kill(c *execdriver.Command, sig int) error {
	logrus.Debugf("jail kill %d %s", sig, c.ID)

	if err := exec.Command("jail", "-r", c.ID).Run(); err != nil {
		return err
	}

	return nil
}

func (d *driver) Pause(c *execdriver.Command) error {
	logrus.Debug("jail pause")
	return nil
}

func (d *driver) Unpause(c *execdriver.Command) error {
	logrus.Debug("jail unpause")
	return nil
}

func (d *driver) Terminate(c *execdriver.Command) error {
	logrus.Debug("jail term")
	return nil
}

func (d *driver) GetPidsForContainer(id string) ([]int, error) {
	logrus.Debugf("jail ps %s", id)
	return nil, nil
}

func (d *driver) Clean(id string) error {
	logrus.Debugf("jail clean %s", id)
	return nil
}

func (d *driver) Exec(c *execdriver.Command, processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes, startCallback execdriver.StartCallback) (int, error) {
	logrus.Debugf("jail exec")
	return 0, nil
}

func (d *driver) Stats(id string) (*execdriver.ResourceStats, error)  {
	logrus.Debugf("jail stats %s", id)
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
	logrus.Debugf("jail isrunning")

	if err := exec.Command("jls", "-j", info.ID).Run(); err != nil {
		return true
	}

	return false
}


// ===

type TtyConsole struct {
	MasterPty *os.File
	SlavePty  *os.File
}

func NewTtyConsole(processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes) (*TtyConsole, error) {
	// lxc is special in that we cannot create the master outside of the container without
	// opening the slave because we have nothing to provide to the cmd.  We have to open both then do
	// the crazy setup on command right now instead of passing the console path to lxc and telling it
	// to open up that console.  we save a couple of openfiles in the native driver because we can do
	// this.
	ptyMaster, ptySlave, err := pty.Open()
	if err != nil {
		return nil, err
	}

	tty := &TtyConsole{
		MasterPty: ptyMaster,
		SlavePty:  ptySlave,
	}

	if err := tty.AttachPipes(&processConfig.Cmd, pipes); err != nil {
		tty.Close()
		return nil, err
	}

	processConfig.Console = tty.SlavePty.Name()

	return tty, nil
}

func (t *TtyConsole) Master() *os.File {
	return t.MasterPty
}

func (t *TtyConsole) Resize(h, w int) error {
	return term.SetWinsize(t.MasterPty.Fd(), &term.Winsize{Height: uint16(h), Width: uint16(w)})
}

func (t *TtyConsole) AttachPipes(command *exec.Cmd, pipes *execdriver.Pipes) error {
	command.Stdout = t.SlavePty
	command.Stderr = t.SlavePty

	go func() {
		if wb, ok := pipes.Stdout.(interface {
			CloseWriters() error
		}); ok {
			defer wb.CloseWriters()
		}

		io.Copy(pipes.Stdout, t.MasterPty)
	}()

	if pipes.Stdin != nil {
		command.Stdin = t.SlavePty
		command.SysProcAttr.Setctty = true

		go func() {
			io.Copy(t.MasterPty, pipes.Stdin)

			pipes.Stdin.Close()
		}()
	}
	return nil
}

func (t *TtyConsole) Close() error {
	t.SlavePty.Close()
	return t.MasterPty.Close()
}