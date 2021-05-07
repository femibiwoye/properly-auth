package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testAddTenant(t *testing.T, ExpectedCode int, phone, businessname, email string) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/tenant/property/add/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["propertyid"] = propertyID[0]
	data["businessname"] = businessname
	data["email"] = email
	data["phone"] = phone
	data["name"] = "abraham akerele"
	data["rentenddate"] = "12/2/2022"
	data["rentstartdate"] = "12/2/2021"

	dataByte, _ := json.Marshal(data)
	mrc := mockReadCloser{data: dataByte}
	req.Body = mrc
	if err != nil {
		t.Fatalf("%v occured", err)
	}
	router.ServeHTTP(w, req)
	responseText, err := ioutil.ReadAll(w.Body)
	if w.Code != ExpectedCode {
		fmt.Printf("%s %s", responseText, w.Result().Status)
		t.Fatalf("Expecting %d Got %d ", ExpectedCode, w.Code)
	}
}

func testRemoveTenant(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/v1/tenant/property/remove/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["propertyid"] = propertyID[0]
	data["userid"] = getIdFromToken(t, tokens[2])

	dataByte, _ := json.Marshal(data)
	mrc := mockReadCloser{data: dataByte}
	req.Body = mrc
	if err != nil {
		t.Fatalf("%v occured", err)
	}
	router.ServeHTTP(w, req)
	responseText, err := ioutil.ReadAll(w.Body)
	if w.Code != ExpectedCode {
		fmt.Printf("%s %s", responseText, w.Result().Status)
		t.Fatalf("Expecting %d Got %d ", ExpectedCode, w.Code)
	}
}

func testListTenant(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/tenant/property/list/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["propertyid"] = propertyID[0]

	dataByte, _ := json.Marshal(data)
	mrc := mockReadCloser{data: dataByte}
	req.Body = mrc
	if err != nil {
		t.Fatalf("%v occured", err)
	}
	router.ServeHTTP(w, req)
	responseText, err := ioutil.ReadAll(w.Body)
	if w.Code != ExpectedCode {
		fmt.Printf("%s %s", responseText, w.Result().Status)
		t.Fatalf("Expecting %d Got %d ", ExpectedCode, w.Code)
	}

	fmt.Printf("%s %s", responseText, w.Result().Status)
}
