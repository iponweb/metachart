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
	"github.com/iponweb/metachart/pkg/chart"
	"github.com/spf13/cobra"
)

type GenCommand struct {
	cmd *cobra.Command
}

func NewGenCommand() *GenCommand {
	command := &GenCommand{}
	cmd := &cobra.Command{
		Use:          "generate",
		Aliases:      []string{"gen"},
		Short:        "generate json-schema and templates using provided configuration",
		Example:      "generate -r .",
		RunE:         command.Execute,
		SilenceUsage: true,
	}

	AddFlags(cmd.Flags())
	command.cmd = cmd
	return command
}

// Register provides cobra.Command initializing and its safe use.
func (command *GenCommand) Register() *cobra.Command {
	return command.cmd
}

func (command *GenCommand) Execute(_ *cobra.Command, args []string) (err error) {
	c, err := chart.NewChart(chartRoot)
	if err != nil {
		return err
	}

	err = c.CleanupTemplates()
	if err != nil {
		return err
	}

	err = command.GenTemplates(*c)
	if err != nil {
		return err
	}

	err = command.GenSchema(*c)
	if err != nil {
		return err
	}

	err = command.GenDocs(*c)
	if err != nil {
		return err
	}

	return c.WriteGen()
}
