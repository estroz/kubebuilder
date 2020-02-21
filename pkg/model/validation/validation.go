/*
Copyright 2018 The Kubernetes Authors.

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

package validation

import (
	"errors"
	"fmt"
	"regexp"
)

// The following code was copied from "k8s.io/apimachinery/pkg/util/validation"
// to avoid package dependencies. In case of additional functionality from
// "k8s.io/apimachinery" is needed, re-consider whether to add the dependency.
// ---------------------------------------------------------------------------

const (
	dns1123LabelFmt       string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	dns1123LabelMaxLength int    = 56 // = 63 - len("-system")

	dns1123SubdomainFmt      string = dns1123LabelFmt + "(\\." + dns1123LabelFmt + ")*"
	dns1123SubdomainErrorMsg string = "a DNS-1123 subdomain must consist of lower case alphanumeric characters, " +
		"'-' or '.', and must start and end with an alphanumeric character"
	// dns1123SubdomainMaxLength is a subdomain's max length in DNS (RFC 1123)
	dns1123SubdomainMaxLength int = 253
)

var (
	dns1123LabelRegexp     = regexp.MustCompile("^" + dns1123LabelFmt + "$")
	dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")
)

// IsDNS1123Subdomain tests for a string that conforms to the definition of a
// subdomain in DNS (RFC 1123).
func IsDNS1123Subdomain(value string) []string {
	var errs []string
	if len(value) > dns1123SubdomainMaxLength {
		errs = append(errs, maxLenError(dns1123SubdomainMaxLength))
	}
	if !dns1123SubdomainRegexp.MatchString(value) {
		errs = append(errs, regexError(dns1123SubdomainErrorMsg, dns1123SubdomainFmt, "example.com"))
	}
	return errs
}

//IsDNS1123Label tests for a string that conforms to the definition of a label in DNS (RFC 1123).
func IsDNS1123Label(value string) []string {
	var errs []string
	if len(value) > dns1123LabelMaxLength {
		errs = append(errs, maxLenError(dns1123LabelMaxLength))
	}
	if !dns1123LabelRegexp.MatchString(value) {
		errs = append(errs, regexError("invalid value for project name", dns1123LabelFmt))
	}
	return errs
}

// regexError returns a string explanation of a regex validation failure.
func regexError(msg string, fmt string, examples ...string) string {
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

// maxLenError returns a string explanation of a "string too long" validation
// failure.
func maxLenError(length int) string {
	return fmt.Sprintf("must be no more than %d characters", length)
}

// End copied code.
// ---------------------------------------------------------------------------

const (
	// projectVersionFmt defines the project version format from a project config.
	projectVersionFmt string = "[1-9][0-9]*(-(alpha|beta))?"
)

var (
	projectVersionRe = regexp.MustCompile("^" + projectVersionFmt + "$")
)

// ValidateProjectVersion ensures version adheres to the project version format.
func ValidateProjectVersion(version string) error {
	if version == "" {
		return errors.New("project version is empty")
	}
	if !projectVersionRe.MatchString(version) {
		return errors.New(regexError("invalid value for project version", projectVersionFmt))
	}
	return nil
}
