#!/bin/bash -xe
#
# Copyright 2022 IPONWEB
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

functions="$(dirname "$0")/env.sh"
if [ -f "$functions" ]; then
  # shellcheck disable=SC1090
  source "$functions"
fi

# shellcheck disable=SC2046
go build -ldflags "-X main.appVersion=${VERSION} -w -s" -o "bin/${APP_BIN}" cmd/main.go
