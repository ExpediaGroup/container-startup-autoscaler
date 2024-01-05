/*
Copyright 2024 Expedia Group, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func cmdRun(cmd *exec.Cmd, info string, coreErrMsg string, fatalOnErr bool, suppressInfo ...bool) (string, error) {
	suppress := false
	if len(suppressInfo) > 0 && suppressInfo[0] {
		suppress = true
	}

	if info != "" && !suppress {
		fmt.Println(info)
	}

	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		trimmedOutput := strings.Trim(string(combinedOutput), "\n")
		wrappedErr := errors.Wrapf(err, "%s (output: %s)", coreErrMsg, trimmedOutput)

		if fatalOnErr {
			fmt.Println(wrappedErr)
			os.Exit(1)
		}
		return trimmedOutput, wrappedErr
	}

	return strings.Trim(string(combinedOutput), "\n"), nil
}
