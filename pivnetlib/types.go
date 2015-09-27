package pivnetlib

import (
	"io/ioutil"
	"strings"
)

//
// Constants
//
const pivnetTokenFilePath = "/home/ubuntu/.pivnet_token"
const urlPrefix = "https://network.pivotal.io"
const bDebug = false

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
