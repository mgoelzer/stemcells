package pivnetlib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-basic/go-curl"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const pivnetTokenFilePath = "/home/ubuntu/.pivnet_token"
const urlPrefix = "https://network.pivotal.io"

//
// PivNet JSON types
//
type ProductFileInner struct {
	AwsObjectKey       string   `json:"aws_object_key"`
	Description        string   `json:"description"`
	DocsUrl            string   `json:"docs_url"`
	FileType           string   `json:"file_type"`
	FileVersion        string   `json:"file_version"`
	IncludedFiles      []string `json:"included_files"`
	Md5                string   `json:"md5"`
	Name               string   `json:"name"`
	Platforms          []string `json:"platforms"`
	ReleasedAt         string   `json:"released_at,omitempty"` // must be MM/DD/YYYY
	Size               int64    `json:"size,omitempty"`        // removed from pivnet examples?
	SystemRequirements []string `json:"system_requirements"`
}

type ProductFile struct {
	ProductFileInner ProductFileInner `json:"product_file"`
}

type ReleaseInner struct {
	Version               string            `json:"version"`
	ReleaseNotesUrl       string            `json:"release_notes_url"`
	Description           string            `json:"description"`
	ReleaseDate           string            `json:"release_date"`
	ReleaseType           string            `json:"release_type"`
	EndOfSupportDate      string            `json:"end_of_support_date"`
	EndOfGuidanceDate     string            `json:"end_of_guidance_date"`
	EndOfAvailabilityDate string            `json:"end_of_availability_date"`
	Availability          string            `json:"availability"`
	Eula                  map[string]string `json:"eula"`
	OssCompliant          string            `json:"oss_compliant"`
	Eccn                  string            `json:"eccn"`
	LicenseException      string            `json:"license_exception"`
	Controlled            bool              `json:"controlled"`
}

type Release struct {
	ReleaseInner ReleaseInner `json:"release"`
}

//
// Read PivNet token
//
func getPivNetToken() (string, error) {
	fileArr, err := ioutil.ReadFile(pivnetTokenFilePath)
	if err != nil {
		return "", err
	}
	fileContents := string(fileArr)
	fileContents = strings.Trim(fileContents, " \n\r")
	//fmt.Printf("%v,%s\n", len(fileContents), fileContents)
	return fileContents, nil
}

func check(e error) {
	if e != nil {
		fmt.Printf("\n\n")
		panic(e)
	}
}

const bDebug = false

//
// Hits Pivnet authentication verification endpoint to verify everything is working
//
func GetAuthentication() (responseHeaders string, responseBodyJsonObj interface{}, errRet error) {
	// Read the pivnet token
	pivnetToken, err := getPivNetToken()
	if err != nil {
		errRet = err
		return
	}

	easy := curl.EasyInit()
	defer easy.Cleanup()

	// set the url
	endpointUrl := fmt.Sprintf("%v/api/v2/authentication", urlPrefix)
	easy.Setopt(curl.OPT_URL, endpointUrl)
	easy.Setopt(curl.OPT_VERBOSE, false)
	//fmt.Printf("DEBUG:  endpointUrl='%v'\n", endpointUrl)

	// set the pivnet headers
	addPivNetHttpHeaders(easy, pivnetToken)

	// set function to collect response data into string buffer
	response := ""
	fWriteToString := func(buf []byte, userdata interface{}) bool {
		if bDebug {
			fmt.Printf("authentication response> %s", string(buf))
		}
		response += string(buf)
		return true
	}
	easy.Setopt(curl.OPT_WRITEFUNCTION, fWriteToString)

	// invoke curl
	if err := easy.Perform(); err != nil {
		fmt.Printf("curl failed\n")
		errRet = err
		return
	}

	// check the http response status
	statusCodeMsg, responseHeaders, responseBodyJsonObj, err := checkHttpResponse(response)
	if err != nil {
		fmt.Printf("checkHttpResponse returned err=%v\n", err)
		errRet = err
		return
	} else {
		if bDebug {
			fmt.Printf("GetAuthentication success:  %v\n", statusCodeMsg)
		}
		return
	}

}

func checkHttpResponse(response string) (statusCodeMsg string, responseHeaders string, responseBodyJsonObj interface{}, errRet error) {
	// Find status line and return nil for 2xx
	var responseLines []string
	responseLines = strings.Split(response, "\n")
	var responseHeaderLines []string
	var responseBodyLines []string
	bInHeaders := true
	for _, line := range responseLines {
		line = strings.Trim(line, " \n\r")
		if bInHeaders {
			responseHeaderLines = append(responseHeaderLines, line)
		} else {
			responseBodyLines = append(responseBodyLines, line)
		}
		if line == "" {
			bInHeaders = false
		}

		if strings.HasPrefix(line, "Status:") {
			if bDebug {
				fmt.Println("checkHttpResponse>>" + line)
			}
			r, _ := regexp.Compile("Status: ([0-9]+ .*)")
			line = strings.Trim(line, " \r\n")
			if r.MatchString(line) {
				arr := r.FindStringSubmatch(line)
				statusCodeMsg = arr[1]
				if strings.HasPrefix(statusCodeMsg, "20") {
					errRet = nil
				} else {
					errRet = errors.New(fmt.Sprintf("Failed ('Status: %v')\n", statusCodeMsg))
				}
			}
		}
	}
	responseHeaders = strings.Join(responseHeaderLines, "\n")

	responseBodyStr := strings.Join(responseBodyLines, "\n")
	responseBodyStr = strings.Trim(responseBodyStr, " \n\r")
	if bDebug {
		fmt.Printf("\n--in:checkHttpResponse--")
		fmt.Printf("\n----responseBodyStr-----")
		fmt.Printf("\n%v\n", responseBodyStr)
		fmt.Printf("------------------------\n")
	}
	if responseBodyStr != "" {
		responseBodyByteArr := []byte(responseBodyStr)
		err := json.Unmarshal(responseBodyByteArr, &responseBodyJsonObj)
		if err != nil {
			errRet = err
		}
		if bDebug {
			fmt.Printf("Dumping 'responseBodyJsonObj':\n")
			dumpArbitraryJsonObject(responseBodyJsonObj, "")
			fmt.Printf("/dumping 'responseBodyJsonObj'\n")
		}
	}

	return
}

func dumpArbitraryJsonObject(responseBodyJsonObj interface{}, indent string) {
	if indent == "" {
		fmt.Printf("\n")
	}
	m := responseBodyJsonObj.(map[string]interface{})
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			fmt.Printf("%s'%v' is string '%v'\n", indent, k, vv)
		case int:
			fmt.Printf("%s'%v' is int '%v'\n", indent, k, vv)
		case float64:
			fmt.Printf("%s'%v' is float64 '%v'\n", indent, k, vv)
		case []interface{}:
			fmt.Printf("%s'%v' is an array:\n", indent, k)
			for i, u := range vv {
				fmt.Printf("%s  [%v] '%v'\n", indent, i, u)
			}
		case map[string]interface{}:
			fmt.Printf("%s'%v' is recursive type\n", indent, k)
			dumpArbitraryJsonObject(vv, indent+"  ")
		default:
			fmt.Printf("%s'%v' is complex type (%T)\n", indent, k, vv)
		}
	}
}

func CreateRelease(productSlug string, version string, description string) (releaseId int, responseHeaders string, responseBodyJsonObj interface{}, errRet error) {
	// Read the pivnet token
	pivnetToken, err := getPivNetToken()
	if err != nil {
		errRet = err
		return
	}

	easy := curl.EasyInit()
	defer easy.Cleanup()

	// set the url
	endpointUrl := fmt.Sprintf("%v/api/v2/products/%v/releases", urlPrefix, productSlug)
	easy.Setopt(curl.OPT_URL, endpointUrl)
	easy.Setopt(curl.OPT_VERBOSE, false)
	//fmt.Printf("DEBUG:  endpointUrl='%v'\n", endpointUrl)

	// set the pivnet headers
	addPivNetHttpHeaders(easy, pivnetToken)

	// get the post data
	t := time.Now()
	tPlusThreeYears := t.AddDate(3, 0, 0)

	//postData := getCreateReleasePostData("1000", "", t, tPlusThreeYears, tPlusThreeYears, tPlusThreeYears)

	r := &Release{
		ReleaseInner: ReleaseInner{
			Version:               version,
			ReleaseNotesUrl:       "http://docs.pivotal.io",
			Description:           description,
			ReleaseDate:           t.Format("2006-01-02"),
			ReleaseType:           "Minor Release",
			EndOfSupportDate:      tPlusThreeYears.Format("2006-01-02"),
			EndOfGuidanceDate:     tPlusThreeYears.Format("2006-01-02"),
			EndOfAvailabilityDate: tPlusThreeYears.Format("2006-01-02"),
			Availability:          "Admins Only",
			Eula:                  map[string]string{"slug": "pivotal_software_eula"},
			OssCompliant:          "confirm",
			Eccn:                  "5D002",
			LicenseException:      "ENC Unrestricted",
			Controlled:            true,
		},
	}
	postData, err := json.MarshalIndent(r, "", "    ")
	fmt.Printf("\n---POST DATA---\n%s\n---------------\n", postData)
	easy.Setopt(curl.OPT_POST, true)
	easy.Setopt(curl.OPT_POSTFIELDSIZE, len(postData))
	sent := false
	fReadRequest := func(ptr []byte, userdata interface{}) int {
		if !sent {
			sent = true
			ret := copy(ptr, postData)
			return ret
		}
		return 0
	}
	easy.Setopt(curl.OPT_READFUNCTION, fReadRequest)

	// set function to collect response data into string buffer
	response := ""
	fWriteToString := func(buf []byte, userdata interface{}) bool {
		if bDebug {
			fmt.Printf("CreateRelease response> %s", string(buf))
		}
		response += string(buf)
		return true
	}
	easy.Setopt(curl.OPT_WRITEFUNCTION, fWriteToString)

	// invoke curl
	if err := easy.Perform(); err != nil {
		fmt.Printf("curl failed\n")
		errRet = err
		return
	}

	// check the http response status
	statusCodeMsg, responseHeaders, responseBodyJsonObj, err := checkHttpResponse(response)
	if err != nil {
		errRet = err
		return
	} else {
		m := responseBodyJsonObj.(map[string]interface{})
		mRelease := m["release"].(map[string]interface{})
		//fmt.Printf("***mRelease = '%v'\n", mRelease)
		idFloat64 := mRelease["id"].(float64)
		releaseId = int(idFloat64)
		//fmt.Printf("releaseId = '%v'\n", releaseId)
		//fmt.Printf("idFloat64 = '%v'\n", idFloat64)
		if bDebug {
			fmt.Printf("CreateRelease success:  %v\n", statusCodeMsg)
		}
		return
	}

}

func DeleteRelease(productSlug string, releaseId int) error {
	// Read the pivnet token
	pivnetToken, err := getPivNetToken()
	if err != nil {
		return err
	}

	// Curl does not support DELETE, so we do it the old fashioned way
	client := &http.Client{}
	endpointUrl := fmt.Sprintf("%v/api/v2/products/%v/releases/%v", urlPrefix, productSlug, releaseId)
	b := bytes.NewBufferString("")
	req, err := http.NewRequest("DELETE", endpointUrl, b)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+pivnetToken)
	reply, err := client.Do(req)
	if err != nil {
		return err
	}
	if bDebug {
		fmt.Printf("reply='%v'\n", reply)
	}
	return nil
}

func addPivNetHttpHeaders(easy *curl.CURL, pivnetToken string) {
	easy.Setopt(curl.OPT_HTTPHEADER, []string{
		"Accept: application/json",
		"Content-Type: application/json",
		"Authorization: Token " + pivnetToken})
	easy.Setopt(curl.OPT_HEADER, 1)
}

func CreateProductFile(productSlug string, pivnetHumanFilename string, awsObjectKey string, description string, md5String string, version string, docsUrl string, release_date time.Time) (productFileId int, responseHeaders string, responseBodyJsonObj interface{}, errRet error) {
	// Read the pivnet token
	pivnetToken, err := getPivNetToken()
	if err != nil {
		return -1, "", "", err
	}

	easy := curl.EasyInit()
	defer easy.Cleanup()

	// set the url
	endpointUrl := fmt.Sprintf("%v/api/v2/products/%v/product_files", urlPrefix, productSlug)
	easy.Setopt(curl.OPT_URL, endpointUrl)
	easy.Setopt(curl.OPT_VERBOSE, false)
	//fmt.Printf("DEBUG:  endpointUrl='%v'\n", endpointUrl)

	// set the pivnet headers
	addPivNetHttpHeaders(easy, pivnetToken)

	m := &ProductFile{
		ProductFileInner: ProductFileInner{
			AwsObjectKey:       awsObjectKey,
			Description:        description,
			DocsUrl:            docsUrl,
			FileType:           "Software",
			FileVersion:        version,
			IncludedFiles:      []string{},
			Md5:                md5String,
			Name:               pivnetHumanFilename,
			SystemRequirements: []string{},
		},
	}
	postData, err := json.MarshalIndent(m, "", "    ")
	//fmt.Printf("\n---POST DATA---\n%s\n---------------\n", postData)
	easy.Setopt(curl.OPT_POST, true)
	easy.Setopt(curl.OPT_POSTFIELDSIZE, len(postData))
	sent := false
	fReadRequest := func(ptr []byte, userdata interface{}) int {
		if !sent {
			sent = true
			ret := copy(ptr, postData)
			return ret
		}
		return 0
	}
	easy.Setopt(curl.OPT_READFUNCTION, fReadRequest)

	// set function to collect response data into string buffer
	response := ""
	fWriteToString := func(buf []byte, userdata interface{}) bool {
		if bDebug {
			fmt.Printf("CreateProductFile response> %s", string(buf))
		}
		response += string(buf)
		return true
	}
	easy.Setopt(curl.OPT_WRITEFUNCTION, fWriteToString)

	// invoke curl
	if err := easy.Perform(); err != nil {
		fmt.Printf("curl failed\n")
		errRet = err
		return
	}

	// check the http response status
	statusCodeMsg, responseHeaders, responseBodyJsonObj, err := checkHttpResponse(response)
	if err != nil {
		fmt.Printf("Error:  checkHttpResponse returned err=%v\n", err)
		errRet = err
		return
	} else {
		fmt.Printf(">>>>statusCodeMsg='%v'\n", statusCodeMsg)
		m := responseBodyJsonObj.(map[string]interface{})
		//fmt.Printf("***m = '%v'\n", m)
		mProductFile := m["product_file"].(map[string]interface{})
		//fmt.Printf("***mProductFile = '%v'\n", mProductFile)
		idFloat64 := mProductFile["id"].(float64)
		productFileId = int(idFloat64)
		//fmt.Printf("productFileId = '%v'\n", productFileId)
		//fmt.Printf("idFloat64 = '%v'\n", idFloat64)
		if bDebug {
			fmt.Printf("CreateProductFile success:  %v\n", statusCodeMsg)
		}
		return
	}
}
