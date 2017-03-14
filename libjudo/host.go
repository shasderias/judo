package libjudo

import (
	"fmt"
	"log"
	"os"
	"path"
)

// Represents a single host (invocation target)
type Host struct {
	Name   string
	Env    map[string]string
	groups []string
	tmpdir string
	cancel chan bool
	master *Proc
	logger *log.Logger
}

func NewHost(name string) (host *Host) {
	env := make(map[string]string)
	env["HOSTNAME"] = name
	return &Host{
		Name:   name,
		Env:    env,
		groups: []string{},
		cancel: make(chan bool),
		master: nil,
		logger: log.New(os.Stderr, fmt.Sprintf("%s: ", name), 0),
	}
}

func (host *Host) SendRemoteAndRun(job *Job) (err error) {
	// speedify!
	host.StartMaster()

	// deferred functions are called first in, last out.
	// any other defers can still use the master to clean up remote.
	defer host.StopMaster()

	// make cozy
	err = host.Ssh(job, "mkdir -p $HOME/.judo")
	tmpdir, err := host.SshRead(job, "TMPDIR=$HOME/.judo mktemp -d")
	if err != nil {
		return err
	}
	host.tmpdir = tmpdir

	cleanup := func() error {
		host.tmpdir = ""
		return host.Ssh(job, fmt.Sprintf("rm -r %s", tmpdir))
	}

	// ensure cleanup
	defer func() {
		if err := recover(); err != nil {
			// oops! clean up remote
			assert(cleanup())
			// continue panicking
			panic(err)
		}
	}()

	// push files to remote
	host.pushFiles(job, job.Script.fname, tmpdir)

	// are we in dirmode?
	var remote_command string
	if !job.Script.dirmode {
		remote_command = path.Join(tmpdir, path.Base(job.Script.fname))
	} else {
		remote_command = path.Join(
			tmpdir,
			path.Base(job.Script.fname),
			"script",
		)
	}

	// do the actual work
	err_job := host.Ssh(job, remote_command)

	// clean up
	if err = cleanup(); err != nil {
		return err
	}
	return err_job
}

func (host *Host) RunRemote(job *Job) (err error) {
	return host.Ssh(job, job.Command.cmd)
}

func (host *Host) Cancel() {
	go func() {
		// kill up to two: master and currently running
		host.cancel <- true
		host.cancel <- true
	}()
}

func (host *Host) StartMaster() (err error) {
	if host.master != nil {
		panic("there already is a master")
	}
	proc, err := NewProc("ssh", "-MN", host.Name)
	if err != nil {
		return
	}
	host.master = proc
	go func() {
		for host.master != nil {
			select {
			case line, ok := <-host.master.Stdout():
				if !ok {
					continue
				}
				host.logger.Println(line)
			case line, ok := <-host.master.Stderr():
				if !ok {
					continue
				}
				host.logger.Println(line)
			case err = <-host.master.Done():
				if err != nil {
					host.logger.Println(err.Error())
				}
				host.master = nil
			case <-host.cancel:
				host.master.CloseStdin()
				host.StopMaster()
			}
		}
	}()
	return
}

func (host *Host) StopMaster() error {
	if host.master == nil || !host.master.IsAlive() {
		host.logger.Println("there was no master to stop")
		return nil
	}
	return host.master.Signal(os.Interrupt)
}
