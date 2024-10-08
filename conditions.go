package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"syscall"
)

/******************************************************************************/
type FileExistsCondition struct {
	FileName string
	Exists   bool
}

func (self FileExistsCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self FileExistsCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Exists {
		return self, self.Exists, nil
	}

	_, err := os.Stat(self.FileName)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return self, false, nil
	}

	if err != nil {
		return self, false, err
	}

	condition := FileExistsCondition{FileName: self.FileName, Exists: true}
	return condition, condition.Exists, nil
}

/******************************************************************************/
type FileRemovedCondition struct {
	FileName string
	Removed  bool
}

func (self FileRemovedCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self FileRemovedCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Removed {
		return self, self.Removed, nil
	}

	_, err := os.Stat(self.FileName)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		condition := FileRemovedCondition{FileName: self.FileName, Removed: true}
		return condition, condition.Removed, nil
	}

	return self, false, err
}

/******************************************************************************/
type FileUpdatedCondition struct {
	FileName string
	FileInfo *fs.FileInfo
	Changed  bool
}

func (self FileUpdatedCondition) Init(ctx *Context) (Condition, error) {
	fileInfo, err := os.Stat(self.FileName)
	if err != nil {
		return self, err
	}

	self.FileInfo = &fileInfo
	return self, nil
}

func (self FileUpdatedCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Changed {
		return self, self.Changed, nil
	}

	fileInfo, err := os.Stat(self.FileName)
	if err != nil {
		return self, false, err
	}

	if !(*self.FileInfo).ModTime().Equal(fileInfo.ModTime()) {
		condition := FileUpdatedCondition{FileName: self.FileName, FileInfo: self.FileInfo, Changed: true}
		return condition, condition.Changed, nil
	}

	return self, self.Changed, nil
}

/******************************************************************************/
type DirExistsCondition struct {
	DirName string
	Exists  bool
}

func (self DirExistsCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self DirExistsCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Exists {
		return self, self.Exists, nil
	}

	fileInfo, err := os.Stat(self.DirName)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return self, false, nil
	}

	if err != nil {
		return self, false, err
	}

	// NB: has to be a directory
	if !fileInfo.Mode().IsDir() {
		return self, self.Exists, nil
	}

	condition := DirExistsCondition{DirName: self.DirName, Exists: true}
	return condition, condition.Exists, nil
}

/******************************************************************************/
type DirRemovedCondition struct {
	DirName string
	Removed bool
}

func (self DirRemovedCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self DirRemovedCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Removed {
		return self, self.Removed, nil
	}

	_, err := os.Stat(self.DirName)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return DirRemovedCondition{DirName: self.DirName, Removed: true}, true, nil
	}

	if err != nil {
		return self, false, err
	}

	// it exists
	return self, false, nil
}

/******************************************************************************/
type DirUpdatedCondition struct {
	DirName  string
	FileInfo *fs.FileInfo
	Changed  bool
}

func (self DirUpdatedCondition) Init(ctx *Context) (Condition, error) {
	fileInfo, err := os.Stat(self.DirName)
	if err != nil {
		return self, err
	}

	self.FileInfo = &fileInfo
	return self, nil
}

func (self DirUpdatedCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Changed {
		return self, self.Changed, nil
	}

	fileInfo, err := os.Stat(self.DirName)
	if err != nil {
		return self, false, err
	}

	if ctx.Verbose {
		fmt.Printf("DirUpdatedCondition: prev:%v curr:%v\n", (*self.FileInfo).ModTime(), fileInfo.ModTime())
	}

	if !(*self.FileInfo).ModTime().Equal(fileInfo.ModTime()) {
		condition := DirUpdatedCondition{DirName: self.DirName, FileInfo: &fileInfo, Changed: true}
		return condition, condition.Changed, nil
	}

	return self, false, nil
}

/******************************************************************************/
type PidExitedCondition struct {
	Pid    int
	Exited bool
}

func (self PidExitedCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self PidExitedCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Exited {
		return self, self.Exited, nil
	}

	pinfo, err := os.FindProcess(self.Pid)
	if err != nil {
		return self, false, err
	}

	// the docs: https://pkg.go.dev/os#Process.Signal
	// don't seem to distinguish the errors returned by Signal, we'll
	// assume that nil means the process exists and non-nil means it
	// does not.
	err = pinfo.Signal(syscall.Signal(0))
	if err == nil {
		return self, false, nil
	}

	return PidExitedCondition{Pid: self.Pid, Exited: true}, true, nil
}

/******************************************************************************/
type CommandExitedCondition struct {
	CommandStr string
	Command    *exec.Cmd
	Exited     bool
	ExitChan   chan error
}

func (self CommandExitedCondition) Init(ctx *Context) (Condition, error) {
	// runtime.Breakpoint()
	cmd := exec.Command("bash", "-c", self.CommandStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()

	if err != nil {
		return self, err
	}

	exitChan := make(chan error, 1)
	go func() {
		if ctx.Verbose {
			fmt.Printf("CommandExitedCondition: START go func: calling cmd.Wait\n")
		}
		res := cmd.Wait()
		exitChan <- res
		if ctx.Verbose {
			fmt.Printf("CommandExitedCondition: EXIT  go func: called cmd.Wait res=%v\n", res)
		}
	}()

	return CommandExitedCondition{CommandStr: self.CommandStr, Command: cmd, Exited: false, ExitChan: exitChan}, nil
}

func (self CommandExitedCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Exited {
		return self, self.Exited, nil
	}

	var err error
	done := false
	select {
	case err = <-self.ExitChan:
		if ctx.Verbose {
			fmt.Printf("CommandExitedCondition: DONE! err=%v\n", err)
		}
		done = true
	default:
		if ctx.Verbose {
			fmt.Printf("CommandExitedCondition: default\n")
		}
	}

	if err != nil {
		return self, false, err
	}

	if done {
		return CommandExitedCondition{CommandStr: self.CommandStr, Exited: true}, true, nil
	}

	// return self, false, nil

	pinfo, err := os.FindProcess(self.Command.Process.Pid)
	if err != nil {
		return self, false, err
	}

	// the docs: https://pkg.go.dev/os#Process.Signal
	// don't seem to distinguish the errors returned by Signal, we'll
	// assume that nil means the process exists and non-nil means it
	// does not.
	err = pinfo.Signal(syscall.Signal(0))
	if err == nil {
		return self, false, nil
	}

	if err != nil && errors.Is(err, os.ErrProcessDone) {
		return CommandExitedCondition{CommandStr: self.CommandStr, Command: nil, Exited: true}, true, nil
	}

	// if err was != nil, assume it means the process has exited, we'll need to clean up
	_, err = pinfo.Wait() // returns ProcessState, error

	if err != nil {
		errno, ok := err.(syscall.Errno)
		if ok {
			if errno == syscall.ECHILD {
				// no child process (we waited for it)
				return CommandExitedCondition{CommandStr: self.CommandStr, Command: nil, Exited: true}, true, nil
			}
		}
	}

	if err != nil {
		return self, false, err
	}

	return CommandExitedCondition{CommandStr: self.CommandStr, Command: nil, Exited: true}, true, nil
}

/******************************************************************************/
type CommandSucceedsCondition struct {
	CommandStr string
}

func (self CommandSucceedsCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self CommandSucceedsCondition) Check(ctx *Context) (Condition, bool, error) {
	cmd := exec.Command("bash", "-c", self.CommandStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return self, false, err
	}

	err = cmd.Wait()
	if err, ok := err.(*exec.ExitError); ok {
		return self, false, err
	}

	return CommandExitedCondition{CommandStr: self.CommandStr, Command: nil, Exited: true}, true, nil
}

/******************************************************************************/
type CommandFailsCondition struct {
	CommandStr string
	Command    *exec.Cmd
	Exited     bool
	ExitChan   chan error
}

func (self CommandFailsCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self CommandFailsCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Exited {
		return self, self.Exited, nil
	}

	var err error
	done := false
	select {
	case err = <-self.ExitChan:
		if ctx.Verbose {
			fmt.Printf("CommandExitedCondition: DONE! err=%v\n", err)
		}
		done = true
	default:
		if ctx.Verbose {
			fmt.Printf("CommandExitedCondition: default\n")
		}
	}

	if err != nil {
		return self, false, err
	}

	if done {
		return CommandExitedCondition{CommandStr: self.CommandStr, Exited: true}, true, nil
	}

	// return self, false, nil

	pinfo, err := os.FindProcess(self.Command.Process.Pid)
	if err != nil {
		return self, false, err
	}

	// the docs: https://pkg.go.dev/os#Process.Signal
	// don't seem to distinguish the errors returned by Signal, we'll
	// assume that nil means the process exists and non-nil means it
	// does not.
	err = pinfo.Signal(syscall.Signal(0))
	if err == nil {
		return self, false, nil
	}

	if err != nil && errors.Is(err, os.ErrProcessDone) {
		return CommandExitedCondition{CommandStr: self.CommandStr, Command: nil, Exited: true}, true, nil
	}

	// if err was != nil, assume it means the process has exited, we'll need to clean up
	_, err = pinfo.Wait() // returns ProcessState, error

	if err != nil {
		errno, ok := err.(syscall.Errno)
		if ok {
			if errno == syscall.ECHILD {
				// no child process (we waited for it)
				return CommandExitedCondition{CommandStr: self.CommandStr, Command: nil, Exited: true}, true, nil
			}
		}
	}

	if err != nil {
		return self, false, err
	}

	return CommandExitedCondition{CommandStr: self.CommandStr, Command: nil, Exited: true}, true, nil
}

/******************************************************************************/
type SocketConnectCondition struct {
	Succeeded bool
	Address   string
}

func (self SocketConnectCondition) Init(ctx *Context) (Condition, error) {
	return self, nil
}

func (self SocketConnectCondition) Check(ctx *Context) (Condition, bool, error) {
	if self.Succeeded {
		return self, self.Succeeded, nil
	}

	dialer := net.Dialer{Timeout: 0}
	conn, err := dialer.Dial("tcp", self.Address)
	if err != nil && errors.Is(err, syscall.ECONNREFUSED) {
		return self, false, nil
	}

	if err != nil {
		return self, false, err
	}

	defer conn.Close()

	return SocketConnectCondition{Succeeded: true, Address: self.Address}, true, nil
}

/******************************************************************************/
// TODO: HttpHeadOkCondition

/******************************************************************************/
// TODO: HttpsHeadOkCondition
