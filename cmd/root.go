package cmd

import (

	//"sigs.k8s.io/kustomize/pkg/commands/build"

	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kelp",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
 examples and usage of using your application. For example:
 
 Cobra is a CLI library for Go that empowers applications.
 This application is a tool to generate the needed files
 to quickly create a Cobra application.`,

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
					templatePath := path.Join(templatesPath, t.Name())
					/*
						1st Stage
						Convert all the golang templating stuff to strings
					*/
					processedTemplate = helmTemplate2KelpTemplate(templatePath)
					/*
						2nd Stage
						Override with kustumization values
					*/
					//processedTemplate = kelpApplyKustomization(processedTemplate)
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

		//fmt.Println("Checking for kustomization file")
		//build.NewOptions("", "")
	},
}

func helmTemplate2KelpTemplate(templatePath string) []string {
	var processedTemplate = []string{}
	template, err := os.Open(templatePath)
	if err != nil {
		log.Fatal(err)
	}
	defer template.Close()

	scanner := bufio.NewScanner(template)
	for scanner.Scan() {
		line := scanner.Text()
		if check, _ := regexp.MatchString("^{{", line); check == true {
		} else if check, _ := regexp.MatchString("\"{{.*(}})?({{)?.*}}\"", line); check == false {
			if check, _ := regexp.MatchString("{{.*(}})?({{)?.*}}", line); check == true {
				line = strings.Replace(line, "{{", "\"{{", -1)
				line = strings.Replace(line, "}}", "}}\"", -1)
			}
		}
		processedTemplate = append(processedTemplate, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return processedTemplate
}
