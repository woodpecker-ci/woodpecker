// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compiler

import backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"

/* cSpell:disable */

var binaryVars = []string{
	"PATH",                         // Specifies directories to search for executable files
	"PATH_SEPARATOR",               // Defines the separator used in the PATH variable
	"COMMAND_MODE",                 // (macOS): Can affect how certain commands are interpreted
	"DYLD_FALLBACK_FRAMEWORK_PATH", // (macOS): Specifies additional locations to search for frameworks
	"DYLD_FALLBACK_LIBRARY_PATH",   // (macOS): Specifies additional locations to search for libraries
}

var libraryVars = []string{
	"LD_PRELOAD",            // Specifies shared libraries to be loaded before all others
	"LD_LIBRARY_PATH",       // Specifies directories to search for shared libraries before the standard locations
	"LD_AUDIT",              // Specifies a shared object to be used for auditing
	"LD_BIND_NOW",           // Forces all relocations to be processed immediately
	"LD_PROFILE",            // Specifies a shared object to be used for profiling
	"LIBPATH",               // (AIX): Similar to LD_LIBRARY_PATH on AIX systems
	"DYLD_INSERT_LIBRARIES", // (macOS): Similar to LD_PRELOAD on macOS
	"DYLD_LIBRARY_PATH",     // (macOS): Similar to LD_LIBRARY_PATH on macOS
}

/* cSpell:enable */

func environmentAllowed(envKey string, stepType backend_types.StepType) bool {
	switch stepType {
	case backend_types.StepTypePlugin,
		backend_types.StepTypeClone:
		for _, v := range append(binaryVars, libraryVars...) {
			if envKey == v {
				return false
			}
		}
	}
	return true
}
