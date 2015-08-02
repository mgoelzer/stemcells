package main

// Must install Amazon Go SDK:  go get -u github.com/aws/aws-sdk-go/...
//   and      codegangsta/cli:  go get -u github.com/codegangsta/cli

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/golang-basic/go-curl"
	"os"
	"strconv"
	"strings"
	"time"

	//	"github.com/aws/aws-sdk-go/aws"
	//	"github.com/aws/aws-sdk-go/aws/awserr"
	//	"github.com/aws/aws-sdk-go/aws/awsutil"
	//	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/codegangsta/cli"
)

const boshIoUrlPrefix = "https://bosh.io/d/stemcells/"

// Urls of the four stemcells we want
type Stemcell struct {
	boshIoName string
	pivNetName string
	version    int
}

var awsStemcell Stemcell = Stemcell{boshIoName: "bosh-aws-xen-hvm-ubuntu-trusty-go_agent", pivNetName: "zzz", version: 3026}

var vsphereStemcell Stemcell = Stemcell{boshIoName: "bosh-vsphere-esxi-ubuntu-trusty-go_agent", pivNetName: "zzz", version: 3026}
var vcdStemcell Stemcell = Stemcell{boshIoName: "bosh-vcloud-esxi-ubuntu-trusty-go_agent", pivNetName: "zzz", version: 3026}
var openstackStemcell Stemcell = Stemcell{boshIoName: "bosh-openstack-kvm-ubuntu-trusty-go_agent-raw", pivNetName: "zzz", version: 3026}

func main() {
	stemcells := make([]Stemcell, 0)
	stemcells = append(stemcells, awsStemcell, vsphereStemcell, vcdStemcell, openstackStemcell)
	app := cli.NewApp()
	app.Name = "stemcell"
	app.Version = "0.1"
	app.Usage = fmt.Sprintf("%s [--help] VERSION", app.Name)
	app.Commands = nil
	app.Copyright = "Copyright (c) 2015 Mike Goelzer (BSD License)"
	cli.AppHelpTemplate = `{{.Name}} {{.Version}} - fetches stemcells for vSphere, vCD, Openstack and AWS
USAGE
  {{.Usage}}
EXAMPLE
  stemcell 3026
{{.Copyright}}
`
	app.Action = func(c *cli.Context) {
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

		for _, value := range stemcells {
			stemcellBoshIoName := value.boshIoName
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
