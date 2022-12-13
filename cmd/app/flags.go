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
	"github.com/spf13/pflag"
	"os"
)

var (
	chartRoot string
)

func getCwdSafe() string {
	result, _ := os.Getwd()
	return result
}

func AddFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&chartRoot, "root", "r",
		getCwdSafe(),
		"Path to the chart root",
	)
}
