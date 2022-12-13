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

package app

import (
	"fmt"
	"github.com/iponweb/metachart/pkg/chart"
	"github.com/spf13/cobra"
)

type InitCommand struct {
	cmd *cobra.Command
}

func NewInitCommand() *InitCommand {
	command := &InitCommand{}
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "generate empty Helm Chart in an empty directory",
		Example:      "init -r .",
		RunE:         command.Execute,
		SilenceUsage: true,
	}

	AddFlags(cmd.Flags())
	command.cmd = cmd
	return command
}

// Register provides cobra.Command initializing and its safe use.
func (command *InitCommand) Register() *cobra.Command {
	return command.cmd
}

func (command *InitCommand) Execute(_ *cobra.Command, args []string) (err error) {
	c := chart.NewChartEmpty(chartRoot)

	empty, err := c.IsEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("init can be performed only in the empty directory")
	}

	return c.WriteInit()
}
