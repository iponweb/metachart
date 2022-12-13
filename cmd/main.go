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
	"runtime"

	"github.com/spf13/cobra"

	"github.com/iponweb/metachart/cmd/app"
)

const (
	appName = "metachart"
	//: env variables

	//: exit codes
	exitOk           = 0
	exitCommandError = 2
)

var (
	appVersion = "dev"

	// Exit function
	Exit = func(code int) {
		os.Exit(code)
	}

	rootCmd = &cobra.Command{
		Use:     appName,
		Short:   "Metachart is a tool to generate Helm Charts using json-schema",
		Example: "metachart version",
	}
	versionCmd = &cobra.Command{
		Use:     "version",
		Example: "version",
		Run: func(cmd *cobra.Command, args []string) {
			version := fmt.Sprintf(
				"%[1]s version: %[2]s, %[3]s/%[4]s %[5]s",
				appName, appVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())
			fmt.Println(version)
		},
	}
)

func exit(err error) {
	if err == nil {
		Exit(exitOk)
		// this return makes sense only for testing, due to
		// there's no real system exit from this function, thus far
		// running in tests it will continue to follow the code sequence.
		return
	}
	Exit(exitCommandError)
}

func init() {
	//: registering local commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(app.NewCommands()...)
}

func main() {
	exit(rootCmd.Execute())
}
