package libjudo

import (
	"os"
	"path"
)

// file/directory to be sent to the remote Host for execution
type Script struct {
	fname   string
	dirmode bool
}

// ad-hoc command to be executed on the remote Host
type Command struct {
	cmd string
}

// a set of Hosts on which to run Scripts/Commands
type Job struct {
	*Inventory
	*Script
	*Command
}

// Holds the result of executing a Job
type JobResult map[*Host]error

func (result *JobResult) Report() (successful, failful int) {
	var success []string
	var failed = make(map[string]error)
	for host := range *result {
		err := (*result)[host]
		if err == nil {
			success = append(success, host.Name)
		} else {
			failed[host.Name] = err
		}
	}
	successful = len(success)
	failful = len(failed)
	if failful > 0 {
		for host := range failed {
			logger.Printf("Failed: %s: %s\n", host, failed[host])
		}
	}
	if successful > 0 {
		logger.Printf("Success: %v\n", success)
	}
	return
}

func NewCommand(cmd string) (command *Command) {
	return &Command{cmd}
}

func NewScript(fname string) (script *Script, err error) {
	script = &Script{fname: fname, dirmode: false}
	stat, err := os.Stat(script.fname)
	if err != nil {
		return nil, err
	}
	// figure out if we should run in dirmode
	if stat.IsDir() {
		stat, err = os.Stat(path.Join(script.fname, "script"))
		if err != nil {
			return nil, err
		}
		script.dirmode = true
	}
	return script, nil
}

func (script *Script) IsDirMode() bool {
	return script.dirmode
}

func NewJob(inventory *Inventory, script *Script, command *Command) (job *Job) {
	return &Job{
		Inventory: inventory,
		Command:   command,
		Script:    script,
	}
}

func (job Job) Log(msg string) {
	logger.Printf("%s\n", msg)
}

func (job *Job) Execute() *JobResult {
	// The heart of judo, run the Job on remote Hosts

	logger.Printf("Running: %v", func() (names []string) {
		// look mama, Go has list comprehensions
		for host := range job.GetHosts() {
			names = append(names, host.Name)
		}
		return
	}())

	// Deliver the results of the job's execution on each Host
	var results = make(map[*Host]chan error)

	// Showtime
	for host := range job.GetHosts() {
		ch := make(chan error)
		results[host] = ch
		go func(host *Host, ch chan error) {
			var err error
			if job.Script != nil {
				err = host.SendRemoteAndRun(job.Script)
			} else if job.Command != nil {
				err = host.RunRemote(job.Command)
			} else {
				panic("Should not happen")
			}
			ch <- err
			close(ch)
		}(host, ch)
	}

	// Stats
	var jobresult JobResult = make(map[*Host]error)
	for host := range results {
		jobresult[host] = <-results[host]
	}
	return &jobresult
}