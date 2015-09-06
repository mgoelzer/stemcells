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

//
// Error check function
//
func check(e error) {
	if e != nil {
		fmt.Printf("\n\n")
		panic(e)
	}
}

//
// Hits Pivnet authentication verification endpoint to verify everything is working
//
const bDebug = false

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
		//if strings.Trim(line, " \n\r") == "" {
		if line == "" {
			bInHeaders = false
		}

		if strings.HasPrefix(line, "Status:") {
			//fmt.Println(">>" + line)
			r, _ := regexp.Compile("Status: ([0-9]+ .*)")
			line = strings.Trim(line, " \r\n")
			if r.MatchString(line) {
				arr := r.FindStringSubmatch(line)
				statusCodeMsg := arr[1]
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
		fmt.Printf("\n-----------------------\n")
		fmt.Printf("%v", responseBodyStr)
		fmt.Printf("\n-----------------------\n")
	}
	if responseBodyStr != "" {
		responseBodyByteArr := []byte(responseBodyStr)
		err := json.Unmarshal(responseBodyByteArr, &responseBodyJsonObj)
		if err != nil {
			errRet = err
		}
		if bDebug {
			dumpArbitraryJsonObject(responseBodyJsonObj, "")
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

func getCreateReleasePostData(version string, release_notes_url string, release_date time.Time, end_of_support_date time.Time, end_of_guidance_date time.Time, end_of_availability_date time.Time) (postData string) {
	// set the POST data string
	postData = `{
  "release": {
`
	postData += "    \"version\": \"1001\",\n"
	postData += "    \"release_notes_url\": \"http://docs.pivotal.io/\",\n"
	postData += "    \"description\": \"\",\n"
	postData += "    \"release_date\": \"2015-09-03\",\n"
	postData += "    \"release_type\": \"Minor Release\",\n"
	postData += "    \"end_of_support_date\": \"2018-12-31\",\n"
	postData += "    \"end_of_guidance_date\": \"2018-12-31\",\n"
	postData += "    \"end_of_availability_date\": \"2018-12-31\",\n"
	postData += `
    "availability": "Admins Only",
    "eula": {
      "slug": "pivotal_software_eula"
    },
    "oss_compliant": "confirm",
    "eccn": "5D002",
    "license_exception": "ENC Unrestricted",
    "controlled": true
  }
}`

	if bDebug {
		fmt.Printf("\n---\nPOST DATA---\n%v\n---------------\n", postData)
	}
	return
}

func CreateRelease(productSlug string) (releaseId int, responseHeaders string, responseBodyJsonObj interface{}, errRet error) {
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
	postData := getCreateReleasePostData("1000", "", t, tPlusThreeYears, tPlusThreeYears, tPlusThreeYears)

	// set function to provide POST data
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

func CreateProduct(productSlug string, pivnetHumanFilename string) error {
	// Read the pivnet token
	pivnetToken, err := getPivNetToken()
	if err != nil {
		return err
	}

	easy := curl.EasyInit()
	defer easy.Cleanup()

	// set the url
	endpointUrl := fmt.Sprintf("%v/api/v2/products/%v/product_files", urlPrefix, productSlug)
	easy.Setopt(curl.OPT_URL, endpointUrl)
	easy.Setopt(curl.OPT_VERBOSE, false)
	fmt.Printf("DEBUG:  endpointUrl='%v'\n", endpointUrl)

	// set the pivnet headers
	addPivNetHttpHeaders(easy, pivnetToken)

	// get the post data
	t := time.Now()
	tPlusThreeYears := t.AddDate(3, 0, 0)
	postData := getCreateReleasePostData("1000", "", t, tPlusThreeYears, tPlusThreeYears, tPlusThreeYears)

	// set function to provide POST data
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
		return err
	}

	// check the http response status
	if statusCodeMsg, _, _, err := checkHttpResponse(response); err != nil {
		return err
	} else {
		if bDebug {
			fmt.Printf("CreateProduct success:  %v\n", statusCodeMsg)
		}
		return nil
	}
}

/*
curl -i -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "Authorization: Token $PIVNET_TOKEN" \
  -XPOST -d '
{
  "product_file": {
    "name": "'"$PIVNET_HUMAN_FILENAME"'",
    "aws_object_key": "product_files/Pivotal-CF/'$FILENAME'",
    "description": "",
    "docs_url": null,
    "file_type": "Software",
    "file_version": "'$RELEASE_VERSION'",
    "included_files": [
        "'"$FILE_INCLUDES"'"
    ],
    "md5": "'$MD5SUM'",
    "platforms": [

    ],
    "released_at": "'$RELEASE_DATE'",
    "size": '$FILE_SIZE',
    "system_requirements": [
        "'"$SYSTEM_REQUIREMENTS"'"
    ],
    "_links": {
      "self": {
        "href": "https://network.pivotal.io/api/v2/products/ops-manager/product_files/130"
      }
    }
  }
}' \
  "$URL_PREFIX/api/v2/products/$PRODUCT_SLUG/product_files"

*/
