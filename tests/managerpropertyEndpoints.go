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
	"time"
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
	documents = append(documents, id["documents"].([]interface{})[0].(string))
	images = append(images, id["images"].([]interface{})[0].(string))
}

func testRemoveAttachment(t *testing.T, ExpectedCode int, typeOf string) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/v1/manager/remove/attachment/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["propertyid"] = propertyID[0]
	if typeOf == "images" {
		data["AttachmentName"] = images[0]
	} else {
		data["AttachmentName"] = documents[0]
	}

	data["AttachmentType"] = typeOf

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

func testAddInspection(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/manager/inspection/schedule/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["propertyid"] = propertyID[0]
	data["text"] = "inspection number 1"
	data["date"] = time.Now().Unix()

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

	result := make(map[string]interface{})
	json.Unmarshal(responseText, &result)
	id := result["data"].(map[string]interface{})
	inspectionID = append(inspectionID, id["id"].(string))
}

func testRemoveInspection(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/v1/manager/inspection/delete/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["inspectionid"] = inspectionID[0]

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

func testUpdateInspection(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/manager/inspection/update/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

	data := make(map[string]interface{})
	data["inspectionid"] = inspectionID[0]

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
