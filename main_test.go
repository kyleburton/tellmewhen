package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"
)

/********************************************************************************/
const TEST_FILE_NAME = "./testing/tmp/TestWaitForFileToExist.txt"
const TEST_DIR_NAME = "./testing/tmp/test-directory"

/********************************************************************************/
func SetupEnsureFile(t *testing.T, fname, contents string) error {
	info, err := os.Stat(fname)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Error: unable to stat file: fname=%s; err=%v", fname, err)
		return err
	}

	if info != nil {
		err = os.Remove(fname)
		if err != nil {
			t.Fatalf("Error: unable to remove file: fname=%s; err=%v", fname, err)
			return err
		}
	}

	err = os.WriteFile(fname, []byte(contents), 0o0644)
	if err != nil {
		t.Fatalf("Error: unable to write file: fname=%s, err=%v", fname, err)
		return err
	}

	return nil
}

func SetupEnsureFileDoesNotExist(t *testing.T, fname string) error {
	var err error

	_, err = os.Stat(fname)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}

	err = os.Remove(fname)
	if err != nil {
		t.Fatalf("Error: unable to remove file: fname=%s; err=%v", fname, err)
		return err
	}

	return nil
}

func SetupEnsureTestDirectory(t *testing.T) error {
	info, err := os.Stat(TEST_DIR_NAME)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Error: unable to stat file: TEST_DIR_NAME=%s; err=%v", TEST_DIR_NAME, err)
		return err
	}

	if info == nil {
		return os.MkdirAll(TEST_DIR_NAME, 0o0755)
	}

	if info.Mode().IsDir() {
		return nil
	}

	return fmt.Errorf("Error: '%s' exists but is not a directory!", TEST_DIR_NAME)
}

func SetupEnsureTestDirectoryDoesNotExist(t *testing.T) error {
	info, err := os.Stat(TEST_DIR_NAME)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	if info != nil && !info.Mode().IsDir() {
		return fmt.Errorf("Error: '%s' exists but is not a directory!", TEST_DIR_NAME)
	}

	return os.RemoveAll(TEST_DIR_NAME)
}

/********************************************************************************/
func TestWaitForFileToExist(t *testing.T) {
	var err error
	var condition Condition = FileExistsCondition{FileName: TEST_FILE_NAME}

	err = SetupEnsureFileDoesNotExist(t, TEST_FILE_NAME)
	if err != nil {
		t.Fatalf("Error: unable to ensure file: TEST_FILE_NAME=%s; err=%v", TEST_FILE_NAME, err)
	}

	// now that the file doesn't exist, the check should be false
	condition, res, err := condition.Check()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("TestWaitForFileToExist failed to check if file exists: TEST_FILE_NAME=%s, err=%v", TEST_FILE_NAME, err)
	}

	if res {
		t.Errorf("TestWaitForFileToExist check failed, file should NOT exist: TEST_FILE_NAME=%s", TEST_FILE_NAME)
	}

	err = SetupEnsureFile(t, TEST_FILE_NAME, "some file contents")
	if err != nil {
		t.Fatalf("TestWaitForFileToExist failed to ensure file exists: TEST_FILE_NAME=%s, err=%v", TEST_FILE_NAME, err)
	}

	// now that the file exists, the check should be true
	condition, res, err = condition.Check() // nolint: staticcheck
	if err != nil {
		t.Fatalf("TestWaitForFileToExist failed to check if file exists: TEST_FILE_NAME=%s, err=%v", TEST_FILE_NAME, err)
	}

	if !res {
		t.Errorf("TestWaitForFileToExist check failed, file SHOULD exist: TEST_FILE_NAME=%s", TEST_FILE_NAME)
	}
}

func TestWaitForFileToNotExist(t *testing.T) {
	var err error
	var condition Condition
	condition = FileRemovedCondition{FileName: TEST_FILE_NAME}
	err = SetupEnsureFile(t, TEST_FILE_NAME, "some file contents")
	if err != nil {
		t.Fatalf("Error: unable to ensure file does not exist: TEST_FILE_NAME=%s; err=%v", TEST_FILE_NAME, err)
	}

	// should not exist
	condition, res, err := condition.Check()
	if err != nil {
		t.Fatalf("Error: error executing FileRemovedCondition check: condition=%v; err=%v", condition, err)
	}

	if res {
		t.Fatalf("Error: FileRemovedCondition should have been false as the file exists condition=%v", condition)
	}

	err = SetupEnsureFileDoesNotExist(t, TEST_FILE_NAME)
	if err != nil {
		t.Fatalf("Error: SetupEnsureFileDoesNotExist for TEST_FILE_NAME=%s; failed=%v", TEST_FILE_NAME, err)
	}

	condition, res, err = condition.Check()
	if err != nil {
		t.Fatalf("Error: error executing FileRemovedCondition check: condition=%v; err=%v", condition, err)
	}

	if !res {
		t.Fatalf("Error: FileRemovedCondition should have been true as the file does not exist condition=%v", condition)
	}
}

func TestFileChanged(t *testing.T) {
	var err error
	var condition Condition
	condition = FileUpdatedCondition{FileName: TEST_FILE_NAME}
	err = SetupEnsureFile(t, TEST_FILE_NAME, "before change")
	if err != nil {
		t.Fatalf("Error: unable to ensure file does not exist: TEST_FILE_NAME=%s; err=%v", TEST_FILE_NAME, err)
	}

	currTime := time.Now().Local()
	oneMinAgo := currTime.Add(time.Duration(-1) * time.Minute)
	err = os.Chtimes(TEST_FILE_NAME, oneMinAgo, oneMinAgo)
	if err != nil {
		t.Fatalf("Error: unable to set create and modification time of TEST_FILE_NAME=%s; err=%v", TEST_FILE_NAME, err)
	}

	condition, err = condition.Init()
	if err != nil {
		t.Fatalf("Error: error initializing check: condition=%v; err=%v", condition, err)
	}

	condition, res, err := condition.Check()
	if err != nil {
		t.Fatalf("Error: error executing FileUpdatedCondition check: condition=%v; err=%v", condition, err)
	}

	if res {
		t.Fatalf("Error: FileUpdatedCondition should have been false as the file hasn't changed yet condition=%v", condition)
	}

	err = os.WriteFile(TEST_FILE_NAME, []byte("after change"), 0o0644)
	if err != nil {
		t.Fatalf("Error: unable to update the contents of file: TEST_FILE_NAME=%s; err=%v", TEST_FILE_NAME, err)
	}

	condition, res, err = condition.Check()
	if err != nil {
		t.Fatalf("Error: error executing FileUpdatedCondition check: condition=%v; err=%v", condition, err)
	}

	if !res {
		t.Fatalf("Error: FileUpdatedCondition should have been true as the file has changed condition=%v", condition)
	}

}

func TestDirExistsCondition(t *testing.T) {
	var err error
	var condition Condition
	condition = DirExistsCondition{DirName: TEST_DIR_NAME}
	err = SetupEnsureTestDirectoryDoesNotExist(t)
	if err != nil {
		t.Fatalf("Error: unable to ensure dir=%s does not exist: err=%v", TEST_DIR_NAME, err)
	}

	condition, res, err := condition.Check()
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if res {
		t.Fatalf("Error: expected condition.Check() to be false!")
	}

	err = SetupEnsureTestDirectory(t)
	if err != nil {
		t.Fatalf("Error: unable to ensure dir=%s exists: err=%v", TEST_DIR_NAME, err)
	}

	condition, res, err = condition.Check() // nolint: staticcheck
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if !res {
		t.Fatalf("Error: expected condition.Check() to be true!")
	}
}

func TestDirRemovedCondition(t *testing.T) {
	var err error
	var condition Condition
	condition = DirRemovedCondition{DirName: TEST_DIR_NAME}
	err = SetupEnsureTestDirectory(t)
	if err != nil {
		t.Fatalf("Error: unable to ensure dir=%s exists: err=%v", TEST_DIR_NAME, err)
	}

	condition, res, err := condition.Check()
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if res {
		t.Fatalf("Error: expected condition.Check() to be false!")
	}

	err = SetupEnsureTestDirectoryDoesNotExist(t)
	if err != nil {
		t.Fatalf("Error: unable to ensure dir=%s does not exist: err=%v", TEST_DIR_NAME, err)
	}

	condition, res, err = condition.Check() // nolint: staticcheck
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if !res {
		t.Fatalf("Error: expected condition.Check() to be true!")
	}
}

func TestDirUpdatedCondition(t *testing.T) {
	var err error
	var condition Condition
	condition = DirUpdatedCondition{DirName: TEST_DIR_NAME}
	err = SetupEnsureTestDirectory(t)
	if err != nil {
		t.Fatalf("Error: unable to ensure dir=%s exists: err=%v", TEST_DIR_NAME, err)
	}

	currTime := time.Now().Local()
	oneMinAgo := currTime.Add(time.Duration(-1) * time.Minute)
	err = os.Chtimes(TEST_DIR_NAME, oneMinAgo, oneMinAgo)
	if err != nil {
		t.Fatalf("Error: unable to set create and modification time of TEST_FILE_NAME=%s; err=%v", TEST_FILE_NAME, err)
	}

	condition, err = condition.Init()
	if err != nil {
		t.Fatalf("Error: failed init DirUpdatedCondition; err=%v", err)
	}

	condition, res, err := condition.Check()
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if res {
		t.Fatalf("Error: expected condition.Check() to be false!")
	}

	err = SetupEnsureFile(t, TEST_DIR_NAME+"/test.txt", "some contents")
	if err != nil {
		t.Fatalf("Error: failed to create file in TEST_DIR_NAME=%s; err=%v", TEST_DIR_NAME, err)
	}

	condition, res, err = condition.Check() // nolint: staticcheck
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if !res {
		t.Fatalf("Error: expected condition.Check() to be true!")
	}
}

func TestPidExitedCondition(t *testing.T) {
	var err error
	var condition Condition
	var sleep_cmd = "sleep"
	var sleep_cmd_args = []string{"3600"}
	cmd := exec.Command(sleep_cmd, sleep_cmd_args...)
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Error: failed to exec/Start sleep_cmd=%s; err=%v", sleep_cmd, err)
	}

	condition = PidExitedCondition{Pid: cmd.Process.Pid}

	condition, res, err := condition.Check()
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if res {
		t.Fatalf("Error: expected condition.Check() to be false! (pid=%d should be running)", cmd.Process.Pid)
	}

	pinfo, err := os.FindProcess(cmd.Process.Pid)
	if err != nil {
		t.Fatalf("Error: failed to get sleep_cmd process info for pid=%d; err=%v", cmd.Process.Pid, err)
	}

	err = pinfo.Kill()
	if err != nil {
		t.Fatalf("Error: error terminaing pid=%d; err=%v", cmd.Process.Pid, err)
	}

	_, err = pinfo.Wait() // returns ProcessState, error
	if err != nil {
		t.Fatalf("Error: error waiting (pinfo.Wait() on pid=%d; err=%v", cmd.Process.Pid, err)
	}

	condition, res, err = condition.Check() // nolint: staticcheck
	if err != nil {
		t.Fatalf("Error: failed to run condition.Check() err=%v", err)
	}

	if !res {
		t.Fatalf("Error: expected condition.Check() to be true! (pid=%d should have terminated)", cmd.Process.Pid)
	}

}

func TestSocketConnectCondition(t *testing.T) {
	var err error
	var res bool
	var condition Condition
	// From: https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers
	//     585 tcp/udp - Previously assigned for use of Internet Message
	//     Access Protocol over TLS/SSL (IMAPS), now deregistered in
	//     favour of port 993.
	address := "localhost:585"
	condition = SocketConnectCondition{Address: address}

	condition, res, err = condition.Check() // nolint: staticcheck
	if err != nil {
		t.Errorf("Error: expected address=%s to not have a listening process, NOTE: this test will fail if a process is listening on that port.", address)
		t.Fatalf("Error: expected connection refused, got error: address=%s; err=%v", address, err)
	}

	if res {
		t.Fatalf("Error: expected connection refsued, succeeded?: address=%s; res=%t", address, res)
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Error: unable to create listening socket for testing SocketConnectCondition: err=%v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	condition = SocketConnectCondition{Address: fmt.Sprintf("localhost:%d", port)}

	condition, res, err = condition.Check() // nolint: staticcheck
	if err != nil {
		t.Fatalf("Error: expected connection refused, got error: address=%s; err=%v", address, err)
	}

	if !res {
		t.Fatalf("Error: expected socket connection Check to succeed, it failed? res=%t", res)
	}
}
