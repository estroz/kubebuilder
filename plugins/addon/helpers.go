package addon

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/gobuffalo/flect"

	"sigs.k8s.io/kubebuilder/pkg/model"
)

// This file gathers functions that are likely to be useful to other
// plugins.  Once we have validated they are used in more than one
// place, we can promote them to a shared location.

type PluginFunc func(u *model.Universe) error

// AddFile adds the specified file to the model.
// If the file exists the function returns false and does not modify the Universe
// If the file does not exist, the function returns true and adds the file to the Universe
// If there is a problem with the file the function returns an error
func AddFile(u *model.Universe, add *model.File) (bool, error) {
	p := add.Path
	if p == "" {
		return false, fmt.Errorf("path must be set")
	}

	for _, f := range u.Files {
		if f.Path == p {
			return false, nil
		}
	}

	u.Files = append(u.Files, add)
	return true, nil
}

// ReplaceFileIfExists replaces the specified file in the model by path
// Returns true iff the file was replaced.
func ReplaceFileIfExists(u *model.Universe, add *model.File) bool {
	p := add.Path
	if p == "" {
		panic("path must be set")
	}

	for i, f := range u.Files {
		if f.Path == p {
			u.Files[i] = add
			return true
		}
	}

	return false
}

// ReplaceFile replaces the specified file in the model by path
// If the file does not exist, it returns an error
func ReplaceFile(u *model.Universe, add *model.File) error {
	found := ReplaceFileIfExists(u, add)
	if !found {
		return fmt.Errorf("file not found %q", add.Path)
	}
	return nil
}

func DefaultTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"title":  strings.Title,
		"lower":  strings.ToLower,
		"plural": flect.Pluralize,
	}
}

func RunTemplate(templateName, templateValue string, data interface{}, funcMap template.FuncMap) (string, error) {
	t, err := template.New(templateName).Funcs(funcMap).Parse(templateValue)
	if err != nil {
		return "", fmt.Errorf("error building template %s: %v", templateName, err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", fmt.Errorf("error rendering template %s: %v", templateName, err)
	}

	return b.String(), nil
}
