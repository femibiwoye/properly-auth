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
	req, err := http.NewRequest("POST", "/signup/", nil)
	req.Header.Add("Content-Type", "application/json")
	data := make(map[string]interface{})
	data["role"] = "landlord"
	data["password"] = password
	data["confirmpassword"] = "testingpassword"
	data["email"] = "abrahamakerele38@gmail.com"
	data["name"] = "Akerele Abraham"

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

	if w.Code >= http.StatusOK && w.Code < 300 {
		return ""
	}

	result := make(map[string]string)
	json.Unmarshal(responseText, &result)

	json.Unmarshal([]byte(result["data"]), &result)

	tokens = append(tokens, result["Token"])

	return tokens[len(tokens)-1]
}

func testResetPassword(t *testing.T, ExpectedCode int, email, platform string) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", fmt.Sprintf("/reset/password/?platform=%s", platform), nil)
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

func testChangePassword(t *testing.T, ExpectedCode int, email, oldPassword string) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/change/password/auth/", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))
	data := make(map[string]interface{})
	data["email"] = email
	data["oldpassword"] = oldPassword
	data["password"] = "newpasswordla"

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

func testSignIn(t *testing.T, ExpectedCode int, email, password string) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/signin/", nil)
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

	result := make(map[string]string)
	json.Unmarshal(responseText, &result)

	json.Unmarshal([]byte(result["data"]), &result)

	tokens = append(tokens, result["Token"])

	return tokens[len(tokens)-1]
}

func testGetProfile(t *testing.T, ExpectedCode int) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/profile/", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

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

func testGeneratePumc(t *testing.T, ExpectedCode int) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/generate/pumc/", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))

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

func testChangePasswordByToken(t *testing.T, ExpectedCode int, email, password, token string) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/change/password/token/", nil)
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
