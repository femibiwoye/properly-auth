package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"properlyauth/utils"
	"testing"
)

func getIdFromToken(t *testing.T, token string) string {
	m, err := utils.DecodeJWTToken(token)
	if err != nil {
		t.Fatalf("Err: %v, can decode token", err)
	}
	return m["user_id"]
}

func testUpdateProperty(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/manager/update/property/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["id"] = propertyID[0]
	data["name"] = "Akerele's house"
	data["type"] = "residential"

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

func testCreateProperty(t *testing.T, ExpectedCode int) {
	file, err := os.Open("image.jpg")
	if err != nil {
		t.Fatalf("%v occured", err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("%v occured", err)
	}
	fi, err := file.Stat()
	if err != nil {
		t.Fatalf("%v occured", err)
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	data := make(map[string]interface{})
	data["address"] = "new post"
	data["name"] = "Abraham's house"
	data["type"] = "Commercial"

	for key, val := range data {
		_ = writer.WriteField(key, val.(string))
	}
	//upload 2 files
	for i := 0; i <= 1; i++ {
		part, err := writer.CreateFormFile("images", fmt.Sprintf("%d%s", i, fi.Name()))
		if err != nil {
			t.Fatalf("%v occured", err)
		}
		part.Write(fileContents)
	}

	//upload 2 docs
	for i := 0; i <= 1; i++ {
		part, err := writer.CreateFormFile("documents", fmt.Sprintf("%d%s", i, fi.Name()))
		if err != nil {
			t.Fatalf("%v occured", err)
		}
		part.Write(fileContents)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("%v occured", err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/manager/create/property/?platform=mobile", body)

	if err != nil {
		t.Fatalf("%v occured", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))
	router.ServeHTTP(w, req)

	responseText, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("%v occured", err)
	}
	if w.Code != ExpectedCode {
		fmt.Printf("%s %s", responseText, w.Result().Status)
		t.Fatalf("Expecting %d Got %d ", ExpectedCode, w.Code)
	}

	result := make(map[string]interface{})
	json.Unmarshal(responseText, &result)
	id := result["data"].(map[string]interface{})
	propertyID = append(propertyID, id["id"].(string))

}

func testAddLandlord(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/landlord/property/add/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["propertyid"] = propertyID[0]
	data["userid"] = getIdFromToken(t, tokens[1])

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

func testRemoveLandlord(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/landlord/property/remove/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["propertyid"] = propertyID[0]
	data["userid"] = getIdFromToken(t, tokens[1])

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

func testAddTenant(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/tenant/property/add/?platform=mobile", nil)
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

func testRemoveTenant(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/tenant/property/remove/?platform=mobile", nil)
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

func testListLandLord(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/landlord/property/list/?platform=mobile", nil)
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
