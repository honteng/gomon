package main

import (
	"os"
	"os/exec"
)

type CommandRunner struct {
	// arguments that will be passed the the commands
	args []string

	task *exec.Cmd
}

func (r *CommandRunner) IsRunning() bool {
	return r.task != nil && r.task.ProcessState != nil && !r.task.ProcessState.Exited()
}

func (r *CommandRunner) buildTask(cmd Command, dir string) *exec.Cmd {
	p := exec.Command(cmd[0], cmd[1:]...)
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	p.Dir = dir
	return p
}

func (r *CommandRunner) Start(cmd Command, args []string, dir string) error {
	r.task = r.buildTask(cmd, dir)
	return r.task.Start()
}

func (r *CommandRunner) Wait(cmd Command, args []string, dir string) error {
	return r.task.Wait()
}

func (r *CommandRunner) Stop() error {
	if r.task != nil && r.task.Process != nil {
		if err := r.task.Process.Kill(); err != nil {
			return err
		}
		_, err := r.task.Process.Wait()
		return err
	}
	return nil
}
