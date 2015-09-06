package main

// Must install codegangsta/cli:  go get -u github.com/codegangsta/cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mgoelzer/stemcells/pivnetlib"

	"github.com/codegangsta/cli"
)

const appHelpTemplate = `{{.Name}} {{.Version}} - deletes a stemcell release from Pivotal Network
{{.Copyright}}

USAGE
  {{.Usage}}

FLAGS
  {{range .Flags}}{{.}}
  {{end}}

EXAMPLE
  delete_release 512
`

const pivnetProductSlug = "stemcells"

func main() {

	app := cli.NewApp()
	app.Name = "delete_release"
	app.Version = "0.1.0"
	app.Usage = fmt.Sprintf("%s [FLAGS] RELEASE_ID", app.Name)
	app.Commands = nil
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "run-tests, t",
			Usage: "whether to run the unit tests",
		},
	}
	cli.AppHelpTemplate = appHelpTemplate

	app.Action = func(c *cli.Context) {
		if c.Bool("run-tests") {
			fmt.Printf("Tests coming soon...\n")
			os.Exit(0)
		}
		if len(c.Args()) != 1 {
			fmt.Printf("Error:  wrong number of arguments (try --help)\n")
			os.Exit(255)
		}
		vArg := c.Args()[0]
		releaseId, err := strconv.Atoi(vArg)
		if (err != nil) || (releaseId <= 0) || (releaseId > 999999) {
			fmt.Printf("Error:  need a numeric pivnet release id (try --help)\n")
			os.Exit(255)
		}

		err = pivnetlib.DeleteRelease(pivnetProductSlug, releaseId)
		if err != nil {
			fmt.Printf("\nERROR: %v\n", err)
			return
		} else {
			fmt.Printf("\nDeleteRelease on %v: ok\n", releaseId)
		}

	}
	app.Run(os.Args)
}
