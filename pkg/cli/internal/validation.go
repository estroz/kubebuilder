/*
Copyright 2020 The Kubernetes Authors.

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

package internal

import (
	"errors"
	"regexp"
)

// projectVersionFmt defines the project version format from a project config.
const projectVersionFmt = "[1-9][0-9]*(-(alpha|beta))?"

var projectVersionRe = regexp.MustCompile("^" + projectVersionFmt + "$")

// ValidateProjectVersion ensures version adheres to the project version format.
func ValidateProjectVersion(version string) error {
	if version == "" {
		return errors.New("project version is empty")
	}
	if !projectVersionRe.MatchString(version) {
		return regexError("invalid value for project version", projectVersionFmt)
	}
	return nil
}

// regexError returns an error containing an explanation of a regex validation
// failure.
func regexError(msg string, fmt string, examples ...string) error {
	return errors.New(buildRegexError(msg, fmt, examples...))
}

// regexError returns a string explanation of a regex validation failure.
func buildRegexError(msg string, fmt string, examples ...string) string {
	if len(examples) == 0 {
		return msg + " (regex used for validation is '" + fmt + "')"
	}
	msg += " (e.g. "
	for i := range examples {
		if i > 0 {
			msg += " or "
		}
		msg += "'" + examples[i] + "', "
	}
	msg += "regex used for validation is '" + fmt + "')"
	return msg
}
