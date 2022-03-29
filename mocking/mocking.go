package mocking

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//DoPOST sends a http request to given url with given body
func DoPOST(url string, body string) error {
	request, err := http.NewRequest("POST", url, ioutil.NopCloser(bytes.NewReader([]byte(body))))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	return nil
}

//GetStringFromDatabase fetches a string by ID from the database, can be overwritten for tests
var GetStringFromDatabase = func(entityId string) string {
	//This is the production code accessing the database, ready a string by its ID and returning it
	return ""
}

//ToUpperCaseFromDatabase fetches something from the database and converts the result to upper case
func ToUpperCaseFromDatabase(input string) string {
	fromDB := GetStringFromDatabase(input) //mocked
	//code under test
	return strings.ToUpper(fromDB)
}
