package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alecthomas/kong"
	// "github.com/posener/complete"
	// "github.com/willabides/kongplete"
)

type Condition interface {
	Init(*Context) (Condition, error)
	Check(*Context) (Condition, bool, error)
}

type Notification interface {
	Notify(*Context) (Notification, bool, error)
}

// //////////////////////////////////////
type WaitableThing int

const (
	Invalid = iota
	WaitOnFileExists
	WaitOnFileRemoved
	WaitOnFileChanged
	WaitOnDirExists
	WaitOnDirRemoved
	WaitOnDirChanged
	WaitOnPidExit
	WaitOnSocketConnect
	WaitOnHttpHeadOk
	WaitOnHttpsHeadOk
)

var WaitableThingToStringTable = map[WaitableThing]string{
	Invalid:             "Invalid",
	WaitOnFileExists:    "WaitOnFileExists",
	WaitOnFileRemoved:   "WaitOnFileRemoved",
	WaitOnFileChanged:   "WaitOnFileChanged",
	WaitOnDirExists:     "WaitOnDirExists",
	WaitOnDirRemoved:    "WaitOnDirRemoved",
	WaitOnDirChanged:    "WaitOnDirChanged",
	WaitOnPidExit:       "WaitOnPidExit",
	WaitOnSocketConnect: "WaitOnSocketConnect",
	WaitOnHttpHeadOk:    "WaitOnHttpHeadOk",
	WaitOnHttpsHeadOk:   "WaitOnHttpsHeadOk",
}

var StringToWaitableThingTable = map[string]WaitableThing{
	"Invalid":             Invalid,
	"WaitOnFileExists":    WaitOnFileExists,
	"WaitOnFileRemoved":   WaitOnFileRemoved,
	"WaitOnFileChanged":   WaitOnFileChanged,
	"WaitOnDirExists":     WaitOnDirExists,
	"WaitOnDirRemoved":    WaitOnDirRemoved,
	"WaitOnDirChanged":    WaitOnDirChanged,
	"WaitOnPidExit":       WaitOnPidExit,
	"WaitOnSocketConnect": WaitOnSocketConnect,
	"WaitOnHttpHeadOk":    WaitOnHttpHeadOk,
	"WaitOnHttpsHeadOk":   WaitOnHttpsHeadOk,
}

func (self WaitableThing) String() string {
	return WaitableThingToStringTable[self]
}

func StringToWaitableThing(str string) WaitableThing {
	return StringToWaitableThingTable[str]
}

// //////////////////////////////////////
type NotificationType int

const (
	NotifyViaCommand = iota
	NotifyViaHttpGet
	NotifyViaHttpPost
)

var NotificationTypeToStringTable = map[NotificationType]string{
	NotifyViaCommand:  "NotifyViaCommand",
	NotifyViaHttpGet:  "NotifyViaHttpGet",
	NotifyViaHttpPost: "NotifyViaHttpPost",
}

var StringToNotificationTypeTable = map[string]NotificationType{
	"NotifyViaCommand":  NotifyViaCommand,
	"NotifyViaHttpGet":  NotifyViaHttpGet,
	"NotifyViaHttpPost": NotifyViaHttpPost,
}

func (self NotificationType) String() string {
	return NotificationTypeToStringTable[self]
}

func StringToNotificationType(str string) NotificationType {
	return StringToNotificationTypeTable[str]
}

/******************************************************************************/
type Context struct {
	Verbose         bool
	TellMeByRunning string
}

func (self *Context) WaitSucceeded(condition Condition) {
	if self.TellMeByRunning != "" {
		cmd := exec.Command("bash", "-c", self.TellMeByRunning)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Context.WaitSucceeded: error executing --tellme-via-running='%s'; err=%v\n", self.TellMeByRunning, err)
			panic(err)
		}

		err = cmd.Wait()
		if err != nil {
			fmt.Printf(": error executing -tellme-via-running='%s'; err=%v\n", self.TellMeByRunning, err)
			panic(err)
		}
	}
}

func (self *Context) Finalize(condition Condition) error {
	if self.TellMeByRunning != "" {
		cmd := exec.Command("bash", "-c", self.TellMeByRunning)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Context.Finalize: error executing -tellme-via-running='%s'; err=%v\n", self.TellMeByRunning, err)
			return err
		}

		err = cmd.Wait()
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Context.Finalize: error: don't know how to notify (no --notify-by-running passed?)")
}

func (self *Context) WaitForCondition(condition Condition) error {
	var err error
	var res bool
	condition, err = condition.Init(self)
	if err != nil {
		return err
	}

	for {
		condition, res, err = condition.Check(self)
		if err != nil {
			return err
		}

		if res {
			return self.Finalize(condition)
		}

		time.Sleep(100 * time.Millisecond)
		fmt.Printf(".")
	}

	return fmt.Errorf("WaitForCondition: terminated the for loop w/o succeeding or failing?")
}

// //////////////////////////////////////////////////////////////////////////////
// File Operations
type FileExistsCmd struct {
	FileName string `help:"the path to the file to look for creation of"`
}

func (self *FileExistsCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(FileExistsCondition{FileName: self.FileName})
}

type FileRemovedCmd struct {
	FileName string `help:"the path to the file to watch for removal of"`
}

func (self *FileRemovedCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(FileRemovedCondition{FileName: self.FileName})
}

type FileUpdatedCmd struct {
	FileName string `help:"the path to the file to watch for update of"`
}

func (self *FileUpdatedCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(FileUpdatedCondition{FileName: self.FileName})
}

// Directory Operations
type DirExistsCmd struct {
	DirName string `help:"the path to the dir to look for creation of"`
}

func (self *DirExistsCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(DirExistsCondition{DirName: self.DirName})
}

type DirRemovedCmd struct {
	DirName string `help:"the path to the dir to watch for removal of"`
}

func (self *DirRemovedCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(DirRemovedCondition{DirName: self.DirName})
}

type DirUpdatedCmd struct {
	DirName string `help:"the path to the dir to watch for update of"`
}

func (self *DirUpdatedCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(DirUpdatedCondition{DirName: self.DirName})
}

// Process Operations
type ProcessExitsCmd struct {
	CommandStr string `required:"" name:"command" help:"the command to execute and wait until it terminates"`
}

func (self *ProcessExitsCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(CommandExitedCondition{CommandStr: self.CommandStr})
}

type ProcessSucceedsCmd struct {
	Command string `help:"the command to execute until it succeeds (exit 0), via bash -c '<command>'."`
}

func (self *ProcessSucceedsCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(CommandSucceedsCondition{CommandStr: self.Command})
}

type ProcessFailsCmd struct {
	Command string `help:"the command to execute until it fails (exit non zero), via bash -c '<command>'."`
}

func (self *ProcessFailsCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(CommandFailsCondition{CommandStr: self.Command})
}

type PidExitsCmd struct {
	Pid int `name:"pid" required:"" help:"the pid of the process to wait to terminate."`
}

func (self *PidExitsCmd) Run(ctx *Context) error {
	return ctx.WaitForCondition(PidExitedCondition{Pid: self.Pid})
}

/******************************************************************************/
var CommandLine struct {
	Verbose         bool   `name:"verbose" optional:"" help:"Be verbose"`
	TellMeByRunning string `name:"notify-by-running" help:"Command to execute to notify of completion."`

	PidExits        PidExitsCmd        `cmd:"" name:"pid-exits" optional:"" help:"Notfiy when a pid has exited (return of exit code success/fail)"`
	ProcessExits    ProcessExitsCmd    `cmd:"" name:"process-exits" optional:"" help:"Notify when a process exits (regardless of exit code sucess/fail)"`
	ProcessSucceeds ProcessSucceedsCmd `cmd:"" name:"process-succeeds" optional:"" help:"Notify when a process succeeds"`
	ProcessFails    ProcessFailsCmd    `cmd:"" name:"process-fails" optional:"" help:"Notify when a process fails"`

	DirUpdated DirUpdatedCmd `cmd:"" name:"dir-updated" optional:"" help:"Notify when a directory has changed."`
	DirExists  DirExistsCmd  `cmd:"" name:"dir-exists" optional:"" help:"Notify when a directory was created."`
	DirRemoved DirRemovedCmd `cmd:"" name:"dir-removed" optional:"" help:"Notify when a directory was removed."`

	FileUpdated FileUpdatedCmd `cmd:"" name:"file-updated" optional:"" help:"Notify when a fileectory has changed."`
	FileExists  FileExistsCmd  `cmd:"" name:"file-exists" optional:"" help:"Notify when a fileectory was created."`
	FileRemoved FileRemovedCmd `cmd:"" name:"file-removed" optional:"" help:"Notify when a fileectory was removed."`

	// TODO: SocketConnectCondition
	// TODO: HttpHeadOkCondition
	// TODO: HttpsHeadOkCondition
}

func main() {
	ctx := kong.Parse(&CommandLine)
	err := ctx.Run(&Context{Verbose: CommandLine.Verbose, TellMeByRunning: CommandLine.TellMeByRunning})

	// NB: a method of notificaiton is required
	// --notify-by-running=<CMD> is required

	if err != nil {
		panic(fmt.Errorf("Execution Error: %w", err))
	}

	os.Exit(0)
}
