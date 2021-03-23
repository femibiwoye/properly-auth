package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testSignUp(t *testing.T, ExpectedCode int, password, email string) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/signup/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	data := make(map[string]interface{})
	data["role"] = "landlord"
	data["password"] = password
	data["confirmpassword"] = password
	data["email"] = email
	data["firstname"] = "Abraham"
	data["lastname"] = "Akerele"

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
	token := result["data"].(map[string]interface{})
	tokens = append(tokens, token["token"].(string))

	return tokens[len(tokens)-1]
}

func testResetPassword(t *testing.T, ExpectedCode int, email, platform string) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", fmt.Sprintf("/v1/reset/password/?platform=%s", platform), nil)
	req.Header.Add("Content-Type", "application/json")
	data := make(map[string]interface{})
	data["email"] = email

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

func testChangePassword(t *testing.T, ExpectedCode int, email, oldPassword, newPassword string) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/v1/user/change-password/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))
	data := make(map[string]interface{})
	data["email"] = email
	data["oldpassword"] = oldPassword
	data["password"] = newPassword

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

func testSignIn(t *testing.T, ExpectedCode int, password, email string) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/login/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	data := make(map[string]interface{})
	data["email"] = email
	data["password"] = password

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

	if w.Code >= 400 {
		return ""
	}

	result := make(map[string]interface{})
	json.Unmarshal(responseText, &result)
	token := result["data"].(map[string]interface{})
	tokens = append(tokens, token["token"].(string))

	return tokens[len(tokens)-1]
}

func testGetProfile(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/user/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

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

func testGeneratePumc(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/generate/pumc/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

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

func testChangePasswordByToken(t *testing.T, ExpectedCode int, email, password, token string) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/change/password/token/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")

	data := make(map[string]interface{})
	data["email"] = email
	data["password"] = password
	data["token"] = token

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
