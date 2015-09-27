package pivnetlib

import (
	"encoding/json"
	"fmt"
	"github.com/golang-basic/go-curl"
	"time"
)

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
