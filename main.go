package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Condition interface {
	Init() (Condition, error)
	Check() (Condition, bool, error)
}

type Notification interface {
	Notify() (Notification, bool, error)
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

// //////////////////////////////////////
type Config struct {
	WaitOn             string `json:"wait-on,omitempty"`
	NotifyType         string `json:"notify-via"`
	DoNotify           bool   `json:"notify"`
	NotifyEverySeconds int    `json:"notify-every-seconds"`
	FileName           string `json:"file-name"`
	DirName            string `json:"dir-name"`
	Pid                int    `json:"pid"`
	PidExitCode        int    `json:"pid-exit-code"`
	UseHttps           bool   `json:"use-https"`
	HostOrAddress      string `json:"host-or-address"`
	Port               string `json:"port"`
	NotifyCommand      string `json:"notify-command"`
	NotifyUrl          string `json:"notify-url"`
	// File Operations
	FileExists  string `json:"file-exists"`
	FileRemoved string `json:"file-removed"`
	FileUpdated string `json:"file-updated"`

	// Porocess Operations
	CommandExits string `json:"command-exits"`
	// Notification Options
	TellmeViaRunning string `json:"tellme-via-running"`
}

func (self *Config) ToJson() (string, error) {
	b, err := json.Marshal(self)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (self *Config) FromJson(b []byte) error {
	return json.Unmarshal(b, self)
}

func Help() {
	fmt.Printf("Tell Me When something has happend.\n")
	fmt.Printf("\n")
	fmt.Printf("  What are some \"THINGs that can happen\"?\n")
	fmt.Printf("    File THINGs\n")
	fmt.Printf("       tellmewhen -file-exists=fname\n")
	fmt.Printf("       tellmewhen -file-removed=fname\n")
	fmt.Printf("       tellmewhen -file-updated=fname\n")
	fmt.Printf("    Directory THINGs\n")
	fmt.Printf("       tellmewhen -dir-exists=fname\n")
	fmt.Printf("       tellmewhen -dir-removed=fname\n")
	fmt.Printf("       tellmewhen -dir-updated=fname\n")
	fmt.Printf("\n")
	fmt.Printf("  Process THINGs\n")
	fmt.Printf("     tellmewhen -pid-exits=PID\n")
	fmt.Printf("     tellmewhen -command-exits='CMD'\n")
	fmt.Printf("     tellmewhen -command-exits='CMD' -with-status=NUM\n")
	fmt.Printf("     tellmewhen -command-succeeds='CMD'\n")
	fmt.Printf("     tellmewhen -command-fails='CMD'\n")
	fmt.Printf("\n")
	fmt.Printf("  Socket THINGs\n")
	fmt.Printf("     tellmewhen -socket-can-connect=<net.Dial address>\n")
	fmt.Printf("                -socket-can-connect=<hostname>:<protocol>\n")
	fmt.Printf("                -socket-can-connect=<hostname>:<int-port>\n")
	fmt.Printf("                -socket-can-connect=<ipv4-addr>:<int-port>\n")
	fmt.Printf("                -socket-can-connect=<ipv6-addr>:<int-port>\n")
	fmt.Printf("     tellmewhen -http-head-ok=HOST-PORT\n")
	fmt.Printf("     tellmewhen -https-head-ok=HOST-PORT\n")
	fmt.Printf("\n")
	fmt.Printf("  How can I be told what happened?\n")
	fmt.Printf("     tellmewhen ...wait-on-something... -tellme-via-running='CMD'\n")
	fmt.Printf("     tellmewhen ...wait-on-something... -tellme-via-http-get='URL'\n")
	fmt.Printf("     tellmewhen ...wait-on-something... -tellme-via-http-post='URL' -post-body='FNAME'\n")
	fmt.Printf("  If you'd like to know periodically that THING is \"not done\":\n")
	fmt.Printf("     tellmewhen ...wait-on-something... -update-every=NUM-SECONDS\n")
	fmt.Printf("\n")
	fmt.Printf("EXAMPLES\n")
	fmt.Printf("\n")
	fmt.Printf("     tellmewhen -command-exits=\"curl -o https://someho.st/some-large-package.bz2\" -tellme-via-running=\"zenity -info 'done'\"\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
}

func (self Config) Finalize(condition Condition) error {
	if self.TellmeViaRunning != "" {
		cmd := exec.Command("bash", "-c", self.TellmeViaRunning)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Config.Finalize: error executing -tellme-via-running='%s'; err=%v\n", self.TellmeViaRunning, err)
			return err
		}

		for {
			err := cmd.Wait()
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (self Config) WaitForCondition(condition Condition) error {
	var err error
	var res bool
	for {
		condition, res, err = condition.Check()
		if err != nil {
			fmt.Printf("WaitForCondition: error=%v\n", err)
			return err
		}

		if res {
			return self.Finalize(condition)
		}

		time.Sleep(100 * time.Millisecond)
		fmt.Printf(".")
	}
}

func main() {
	var err error
	AppConfig := Config{}
	printConfig := false
	showHelp := false
	configFile := ""

	// flag.BoolVar(&AppConfig.DoNotify, "do-notify", true, "Notify of progress.")
	flag.IntVar(&AppConfig.NotifyEverySeconds, "update-every", 60, "Elapsed time between notifications.")
	flag.StringVar(&configFile, "config", "", "Load Configuration")
	flag.BoolVar(&printConfig, "print-config", false, "Print the configuration")
	flag.BoolVar(&showHelp, "help", false, "Display Help")

	// File Operations
	flag.StringVar(&AppConfig.FileExists, "file-exists", "", "Wait for a file to exist.")
	flag.StringVar(&AppConfig.FileRemoved, "file-removed", "", "Wait for a file to be removed.")
	flag.StringVar(&AppConfig.FileUpdated, "file-updated", "", "Wait for a file to be updated (mtime).")

	// Process Operations
	flag.StringVar(&AppConfig.CommandExits, "command-exits", "", "Specify a command to exec and wait for.")

	// Notification Options
	flag.StringVar(&AppConfig.TellmeViaRunning, "tellme-via-running", "", "Specify a command to exec to 'tell you'.")
	flag.Parse()

	if configFile != "" {
		b, err := os.ReadFile(configFile)
		if err != nil {
			panic(err)
		}

		err = AppConfig.FromJson(b)
		if err != nil {
			panic(err)
		}
	}

	if printConfig {
		json, err := AppConfig.ToJson()
		if err != nil {
			panic(err)
		}
		fmt.Print(json)
		return
	}

	if showHelp {
		Help()
		return
	}

	// TODO: check for conflicting actions (eg: wait for pid / exec process / socket connect)
	var condition Condition

	// TODO: support all of the condition types
	if AppConfig.CommandExits != "" {
		condition = CommandExitedCondition{CommandStr: AppConfig.CommandExits}
	}

	if AppConfig.FileExists != "" {
		condition = FileExistsCondition{FileName: AppConfig.FileExists}
	}

	if nil == condition {
		panic(fmt.Errorf("Error: you must specify something to wait on!"))
	}

	condition, err = condition.Init()
	if err != nil {
		panic(err)
	}

	err = AppConfig.WaitForCondition(condition)
	if err != nil {
		panic(err)
	}
}
