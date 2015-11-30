package main

import "github.com/convox/rack/Godeps/_workspace/src/github.com/codegangsta/cli"

var appFlag = cli.StringFlag{
	Name:  "app, a",
	Usage: "App name. Inferred from current directory if not specified.",
}

var cfTemplateFlag = cli.StringFlag{
	Name:   "template",
	Value:  "",
	Usage:  "File path for custom cloudformation template",
	EnvVar: "TEMPLATE",
}
