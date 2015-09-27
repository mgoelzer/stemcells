package pivnetlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-basic/go-curl"
	"regexp"
	"strings"
)

//
// Hits Pivnet authentication verification endpoint to verify everything is working
//
func VerifyAuthentication() (responseHeaders string, responseBodyJsonObj interface{}, errRet error) {
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

func addPivNetHttpHeaders(easy *curl.CURL, pivnetToken string) {
	easy.Setopt(curl.OPT_HTTPHEADER, []string{
		"Accept: application/json",
		"Content-Type: application/json",
		"Authorization: Token " + pivnetToken})
	easy.Setopt(curl.OPT_HEADER, 1)
}
