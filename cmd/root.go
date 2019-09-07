package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"sigs.k8s.io/kustomize/pkg/resource"

	"github.com/helm/helm/pkg/chart"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/k8sdeps/validator"
	"sigs.k8s.io/kustomize/pkg/commands/build"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/plugins"
	"sigs.k8s.io/kustomize/pkg/resmap"
)

// Chart is a helm package that contains metadata, a default config, zero or more
// optionally parameterizable templates, and zero or more charts (dependencies).
type Chart struct {
	// Metadata is the contents of the Chartfile.
	Metadata *Metadata
	// LocK is the contents of Chart.lock.
	Lock *Lock
	// Templates for this chart.
	Templates []*File
	// Values are default config for this template.
	Values map[string]interface{}
	// Schema is an optional JSON schema for imposing structure on Values
	Schema []byte
	// Files are miscellaneous files in a chart archive,
	// e.g. README, LICENSE, etc.
	Files []*File

	parent       *Chart
	dependencies []*Chart
}

type File struct {
	// Name is the path-like name of the template.
	Name string
	// Data is the template as byte data.
	Data []byte
}

type Dependency struct {
	// Name is the name of the dependency.
	//
	// This must mach the name in the dependency's Chart.yaml.
	Name string `json:"name"`
	// Version is the version (range) of this chart.
	//
	// A lock file will always produce a single version, while a dependency
	// may contain a semantic version range.
	Version string `json:"version,omitempty"`
	// The URL to the repository.
	//
	// Appending `index.yaml` to this string should result in a URL that can be
	// used to fetch the repository index.
	Repository string `json:"repository"`
	// A yaml path that resolves to a boolean, used for enabling/disabling charts (e.g. subchart1.enabled )
	Condition string `json:"condition,omitempty"`
	// Tags can be used to group charts for enabling/disabling together
	Tags []string `json:"tags,omitempty"`
	// Enabled bool determines if chart should be loaded
	Enabled bool `json:"enabled,omitempty"`
	// ImportValues holds the mapping of source values to parent key to be imported. Each item can be a
	// string or pair of child/parent sublist items.
	ImportValues []interface{} `json:"import-values,omitempty"`
	// Alias usable alias to be used for the chart
	Alias string `json:"alias,omitempty"`
}

// Lock is a lock file for dependencies.
//
// It represents the state that the dependencies should be in.
type Lock struct {
	// Genderated is the date the lock file was last generated.
	Generated time.Time `json:"generated"`
	// Digest is a hash of the dependencies in Chart.yaml.
	Digest string `json:"digest"`
	// Dependencies is the list of dependencies that this lock file has locked.
	Dependencies []*Dependency `json:"dependencies"`
}

// Maintainer describes a Chart maintainer.
type Maintainer struct {
	// Name is a user name or organization name
	Name string `json:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `json:"email,omitempty"`
	// URL is an optional URL to an address for the named maintainer
	URL string `json:"url,omitempty"`
}

// Metadata for a Chart file. This models the structure of a Chart.yaml file.
type Metadata struct {
	// The name of the chart
	Name string `json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `json:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `json:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart
	Version string `json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `json:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	Maintainers []*Maintainer `json:"maintainers,omitempty"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty"`
	// The API Version of this chart.
	APIVersion string `json:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `json:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `json:"deprecated,omitempty"`
	// Annotations are additional mappings uninterpreted by Helm,
	// made available for inspection by other applications.
	Annotations map[string]string `json:"annotations,omitempty"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string `json:"kubeVersion,omitempty"`
	// Dependencies are a list of dependencies for a chart.
	Dependencies []*Dependency `json:"dependencies,omitempty"`
	// Specifies the chart type: application or library
	Type string `json:"type,omitempty"`
}

var RootCmd = &cobra.Command{
	Use: "kelp",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			chartPath := path.Join(cwd, args[0])
			templatesPath := path.Join(chartPath, "templates")
			templates, err := ioutil.ReadDir(templatesPath)
			if err != nil {
				log.Fatal(err)
			}

			for _, t := range templates {
				processedTemplate := []string{}
				if strings.HasSuffix(t.Name(), ".yaml") {
					//templatePath := path.Join(templatesPath, t.Name())
					/*
						1st Stage
						Convert all the golang templating stuff to strings
					*/
					//processedTemplate = helmTemplate2KelpTemplate(templatePath)

					/*
						2nd Stage
						Override with kustumization values
					*/
					kelpApplyKustomization(chartPath)
					/*
						3rd Stage
						Convert back to helm template
					*/
				}
				for _, line := range processedTemplate {
					fmt.Println(line)
				}
			}

		} else {
			panic(errors.New("No arguments given"))
		}
	},
}

func helmTemplate2KelpTemplate(tplPath string) []string {
	var processedTemplate = []string{}
	c := chart.Chart{}

	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		log.Fatal(err)
	}

	tpl.Execute(os.Stdout, c)

	return processedTemplate
}

func kelpApplyKustomization(kustPath string) {
	// Only takes root of the kustomize
	k := build.NewOptions(kustPath, "")

	var resourceFactory *resource.Factory

	v := validator.NewKustValidator()
	pf := transformer.NewFactoryImpl()
	rf := resmap.NewFactory(resourceFactory)
	fSys := fs.MakeRealFS()

	pluginConfig := plugins.DefaultPluginConfig()
	pl := plugins.NewLoader(pluginConfig, rf)

	b := bytes.Buffer{}
	fmt.Printf("%#v", &b)

	err := k.RunBuild(os.Stdout, v, fSys, rf, pf, pl)
	if err != nil {
		panic(err)
	}

}
