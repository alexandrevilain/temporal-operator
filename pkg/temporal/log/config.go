// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package log

import (
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"go.temporal.io/server/common/log"
)

// NewLogConfigFromLogSpec creates a new instance of a temporal log config from the provided LogSpec.
func NewSQLConfigFromDatastoreSpec(spec *v1beta1.LogSpec) log.Config {
	if spec == nil {
		return log.Config{
			Stdout: true,
			Level:  "info",
		}
	}

	stdout := true
	if spec.Stdout != nil {
		stdout = *spec.Stdout
	}
	return log.Config{
		Stdout:      stdout,
		Level:       spec.Level,
		OutputFile:  spec.OutputFile,
		Format:      spec.Format,
		Development: spec.Development,
	}
}
