package main

import (
	"errors"
	"os"
	"testing"
)

func TestWaitForFileToExist(t *testing.T) {
	var err error
	fname := "./testing/tmp/TestWaitForFileToExist.txt"
	var condition Condition = FileExistsCondition{FileName: fname}

	res, err := condition.Check()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("TestWaitForFileToExist setup failed, unable to check if file exists: fname=%s, err=%v", fname, err)
	}

	if res {
		err = os.Remove(fname)
		if err != nil {
			t.Fatalf("TestWaitForFileToExist setup failed, unable to remove file: fname=%s, err=%v", fname, err)
		}
	}

	// now that the file doesn't exist, the check should be false
	res, err = condition.Check()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("TestWaitForFileToExist failed to check if file exists: fname=%s, err=%v", fname, err)
	}

	if res {
		t.Errorf("TestWaitForFileToExist check failed, file should NOT exist: fname=%s", fname)
	}

	err = os.WriteFile(fname, []byte("some file contents"), 0o0755)
	if err != nil {
		t.Fatalf("TestWaitForFileToExist failed to create test file: fname=%s, err=%v", fname, err)
	}

	// now that the file exists, the check should be true
	res, err = condition.Check()
	if err != nil {
		t.Fatalf("TestWaitForFileToExist failed to check if file exists: fname=%s, err=%v", fname, err)
	}

	if !res {
		t.Errorf("TestWaitForFileToExist check failed, file SHOULD exist: fname=%s", fname)
	}
}
