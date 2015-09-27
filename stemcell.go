package main

// Must install codegangsta/cli:  go get -u github.com/codegangsta/cli

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/golang-basic/go-curl"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mgoelzer/stemcells/pivnetlib"

	"github.com/codegangsta/cli"
)

const boshIoUrlPrefix = "https://bosh.io/d/stemcells/"

const awsStemcellBoshIoName = "bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
const vsphereStemcellBoshIoName = "bosh-vsphere-esxi-ubuntu-trusty-go_agent"
const vcdStemcellBoshIoName = "bosh-vcloud-esxi-ubuntu-trusty-go_agent"
const openstackStemcellBoshIoName = "bosh-openstack-kvm-ubuntu-trusty-go_agent-raw"

const appHelpTemplate = `{{.Name}} {{.Version}} - fetches stemcells for vSphere, vCD, Openstack and AWS
{{.Copyright}}

USAGE
  {{.Usage}}

FLAGS
  {{range .Flags}}{{.}}
  {{end}}

EXAMPLE
  stemcell 3026
`

const pivnetProductSlug = "stemcells"

/***************************************************************/
// Temporary section
/***************************************************************/

func testCodeToAuth() {
	if _, responseBodyJsonObj, err := pivnetlib.VerifyAuthentication(); err != nil {
		fmt.Printf("\nERROR: %v\n%v\n", err, responseBodyJsonObj)
		return
	} else {
		fmt.Printf("\nGetAuthentication:  ok\n%v\n", responseBodyJsonObj)
	}
}

func testCodeToCreateRelease() {
	if releaseId, _, responseBodyJsonObj, err := pivnetlib.CreateRelease(pivnetProductSlug, "1000", "Description....."); err != nil {
		fmt.Printf("\nERROR: %v\n%v\n", err, responseBodyJsonObj)
		return
	} else {
		fmt.Printf("\nCreateRelease created release Id:  %v\n", releaseId)
	}
}

func testCodeToDeleteRelease() {
	releaseId := 557
	if err := pivnetlib.DeleteRelease(pivnetProductSlug, releaseId); err != nil {
		fmt.Printf("\nERROR: %v\n", err)
		return
	} else {
		fmt.Printf("\nDeleteRelease deleted release Id:  %v\n", releaseId)
	}
}

func testCodeToCreateProductFile() {
	if productId, responseHeaders, responseBodyJsonObj, err := pivnetlib.CreateProductFile("stemcells", "TEST PRODUCT - Ubuntu Trusty Stemcell for AWS", "product_files/Pivotal-CF/light-bosh-stemcell-2840-aws-xen-hvm-ubuntu-trusty-go_agent.tgz", "Test test test", "abcdef432523", "9999A", "http://docs.pivotal.io", time.Now()); err != nil {
		fmt.Printf("\nCreateProductFile ERROR: %v\n%v\n%v\n", err, responseHeaders, responseBodyJsonObj)
	} else {
		fmt.Printf("\nCreateProductFile created product with Id:  %v\n", productId)
	}

}

/***************************************************************/
// End -- Temporary section
/***************************************************************/

func main() {
	testCodeToAuth()

	//testCodeToCreateRelease()
	//testCodeToDeleteRelease()
	os.Exit(1)
	///////////////////////////////

	stemcellNames := append(make([]string, 0, 4),
		awsStemcellBoshIoName, vsphereStemcellBoshIoName, vcdStemcellBoshIoName, openstackStemcellBoshIoName)
	app := cli.NewApp()
	app.Name = "stemcell"
	app.Version = "0.1.0"
	app.Usage = fmt.Sprintf("%s [FLAGS] VERSION", app.Name)
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
		version, err := strconv.Atoi(vArg)
		if (err != nil) || (version <= 0) || (version > 99999) {
			fmt.Printf("Error:  need a numeric argument (try --help)\n")
			os.Exit(255)
		}

		for _, value := range stemcellNames {
			stemcellBoshIoName := value
			stemcellFilename, stemcellBytes, md5, err := fetchStemcell(stemcellBoshIoName, version)
			if err != nil {
				fmt.Println("fetchStemcell failed")
				os.Exit(255)
			}
			fmt.Printf("%v (%v bytes, %v)\n", stemcellFilename, stemcellBytes, md5)
		}
	}
	app.Run(os.Args)
}

func check(e error) {
	if e != nil {
		fmt.Printf("\n\n")
		panic(e)
	}
}

func fetchStemcell(stemcellBoshIoName string, version int) (string, int, string, error) {
	easy := curl.EasyInit()
	defer easy.Cleanup()

	// Set the URL to fetch
	stemcellUrl := fmt.Sprintf("%v?v=%v", stemcellBoshIoName, version)
	//fmt.Println("DEBUG:  " + boshIoUrlPrefix + stemcellUrl)
	easy.Setopt(curl.OPT_URL, boshIoUrlPrefix+stemcellUrl)
	easy.Setopt(curl.OPT_VERBOSE, false)

	// Get the name in "Location:" header without actually redirecting yet
	easy.Setopt(curl.OPT_FOLLOWLOCATION, false)
	fWriteToDevNull := func(buf []byte, userdata interface{}) bool { return true }
	easy.Setopt(curl.OPT_WRITEFUNCTION, fWriteToDevNull)
	err := easy.Perform()
	check(err)
	locationString, err := easy.Getinfo(curl.INFO_REDIRECT_URL)
	check(err)
	//fmt.Printf("DEBUG:  locationString='%v'\n", locationString)

	locationStringParts := strings.Split(locationString.(string), "/")
	locationStringPartsLen := len(locationStringParts)
	stemcellFilename := locationStringParts[locationStringPartsLen-1]
	//fmt.Printf("DEBUG:  stemcellFilename='%v'\n",stemcellFilename)

	// Open the stemcell file for writing (will be in current dir)
	stemcellLocalPath := stemcellFilename
	f, err := os.Create(stemcellLocalPath)
	check(err)
	defer f.Close()

	// Now fetch again with redirect to fetch file
	easy.Setopt(curl.OPT_FOLLOWLOCATION, true)
	bytesWritten := 0

	hash := md5.New()
	fWriteToFile := func(buf []byte, userdata interface{}) bool {
		bytesWritten += len(buf)
		//fmt.Println("DEBUG:  size=>", len(buf))
		f.Write(buf)
		hash.Write(buf)
		//println("data = >", string(buf))
		return true
	}
	easy.Setopt(curl.OPT_WRITEFUNCTION, fWriteToFile)

	// Progress bar
	easy.Setopt(curl.OPT_NOPROGRESS, false)
	started := int64(0)
	easy.Setopt(curl.OPT_PROGRESSFUNCTION, func(dltotal, dlnow, ultotal, ulnow float64, userdata interface{}) bool {
		if started == 0 {
			started = time.Now().Unix()
		}
		fmt.Printf("%v: %3.2f%%, Speed: %.1fKiB/s \r", stemcellFilename, dlnow/dltotal*100, dlnow/1000/float64((time.Now().Unix()-started)))
		return true
	})

	if err := easy.Perform(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return "", 0, "", errors.New("curl failed")
	}

	return stemcellFilename, bytesWritten, fmt.Sprintf("%x", hash.Sum(nil)), nil
}
