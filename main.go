package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Condition interface {
	Check() (bool, error)
}

type Notification interface {
	Notify() (bool, error)
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
	fmt.Printf("       tellmewhen --file-exists=fname\n")
	fmt.Printf("       tellmewhen --file-removed=fname\n")
	fmt.Printf("       tellmewhen --file-updated=fname\n")
	fmt.Printf("    Directory THINGs\n")
	fmt.Printf("       tellmewhen --dir-exists=fname\n")
	fmt.Printf("       tellmewhen --dir-removed=fname\n")
	fmt.Printf("       tellmewhen --dir-updated=fname\n")
	fmt.Printf("\n")
	fmt.Printf("  Process THINGs\n")
	fmt.Printf("     tellmewhen --pid-exits=PID\n")
	fmt.Printf("     tellmewhen --command-exits='CMD'\n")
	fmt.Printf("     tellmewhen --command-exits='CMD' --with-status=NUM\n")
	fmt.Printf("     tellmewhen --command-succeeds='CMD'\n")
	fmt.Printf("     tellmewhen --command-fails='CMD'\n")
	fmt.Printf("\n")
	fmt.Printf("  Socket THINGs\n")
	fmt.Printf("     tellmewhen --socket-can-connect=HOST-PORT\n")
	fmt.Printf("     tellmewhen --http-head-ok=HOST-PORT\n")
	fmt.Printf("     tellmewhen --https-head-ok=HOST-PORT\n")
	fmt.Printf("\n")
	fmt.Printf("  How can I be told what happened?\n")
	fmt.Printf("     tellmewhen ...wait-on-something... --tellme-via-running='CMD'\n")
	fmt.Printf("     tellmewhen ...wait-on-something... --tellme-via-http-get='URL'\n")
	fmt.Printf("     tellmewhen ...wait-on-something... --tellme-via-http-post='URL'\n")
	fmt.Printf("  If you'd like to know periodically that THING is \"not done\":\n")
	fmt.Printf("     tellmewhen ...wait-on-something... --update-every=NUM-SECONDS\n")
}

func main() {
	AppConfig := Config{}
	printConfig := false
	showHelp := false
	configFile := ""

	flag.BoolVar(&AppConfig.DoNotify, "do-notify", true, "Notify of progress.")
	flag.IntVar(&AppConfig.NotifyEverySeconds, "update-every", 60, "Elapsed time between notifications.")
	flag.StringVar(&configFile, "config", "", "Load Configuration")
	flag.BoolVar(&printConfig, "print-config", false, "Print the configuration")
	flag.BoolVar(&showHelp, "help", false, "Display Help")
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

}
