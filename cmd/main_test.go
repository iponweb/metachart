/*
 * Copyright 2022 IPONWEB
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"

	"os"
	"reflect"
	"testing"
)

func issue(t *testing.T, expected, result interface{}) {
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected: `%v`, got: `%v` instead", expected, result)
	}
}

type exitMock struct {
	stack    []int
	lastCall int
}

func (mock *exitMock) exit(exitCode int) {
	mock.stack = append(mock.stack, exitCode)
	mock.lastCall = exitCode
	fmt.Printf("oh no: %d\n", exitCode)
}

func (mock *exitMock) firstCall() int {
	if len(mock.stack) > 0 {
		return mock.stack[0]
	}
	return -1
}

func (mock *exitMock) clear() {
	mock.stack = []int{}
	mock.lastCall = 0
}

func Test_main(t *testing.T) {
	exitMock := exitMock{}
	backupArgs := os.Args
	exitFunc := Exit

	defer func() {
		os.Args = backupArgs
		Exit = exitFunc
	}()
	Exit = exitMock.exit

	t.Run("ok/help", func(in *testing.T) {
		os.Args = []string{appName, "--help"}
		main()
		if exitMock.lastCall != exitOk {
			in.Errorf("exitOk: %v got %v instead", exitOk, exitMock.lastCall)
		}
	})

	t.Run("ok/version", func(in *testing.T) {
		os.Args = []string{appName, "version"}
		main()
		exitMock.clear()
		if exitMock.lastCall != exitOk {
			in.Errorf("exitOk: %v got %v instead", exitOk, exitMock.lastCall)
		}
	})

	t.Run("app-error", func(in *testing.T) {
		os.Args = []string{appName, "test", "error"}
		main()
		if exitMock.firstCall() != exitCommandError {
			in.Errorf("exitCommandError(%v), got: %v", exitCommandError, exitMock.lastCall)
		}
	})
}

func Test_exit(t *testing.T) {
	exitMock := exitMock{}
	exitFunc := Exit
	defer func() {
		Exit = exitFunc
	}()
	Exit = exitMock.exit

	for _, entry := range []struct {
		name     string
		in       error
		expected int
	}{
		{"ok/no-error", nil, exitOk},
		{"ok/command-error", fmt.Errorf("any error"), exitCommandError},
	} {
		t.Run(entry.name, func(in *testing.T) {
			exit(entry.in)
			if exitMock.lastCall != entry.expected {
				issue(in, entry.in, exitMock.lastCall)
			}
		})
	}
}
