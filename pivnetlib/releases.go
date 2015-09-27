package pivnetlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang-basic/go-curl"
	"net/http"
	"time"
)

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
