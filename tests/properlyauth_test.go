package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"properlyauth/database"
	"properlyauth/models"
	"properlyauth/utils"
	"strings"
	"syscall"
	"testing"

	"github.com/joho/godotenv"
)

func handleInterupt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanUpDb()
		os.Exit(1)
	}()
}

func TestProperly(t *testing.T) {

	os.Setenv("HOST", "127.0.0.1:8080")
	os.Setenv("TESTING", "TESTING")
	os.Setenv("DBNAME", "properlytesting")
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatalf("Error loading .env file")
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Can't get current working directory due to error :%v", err)
	}

	dir = strings.TrimSuffix(dir, "tests")
	os.Setenv("ROOTDIR", dir)
	_, err = os.Stat(fmt.Sprintf("%spublic/media", dir))
	if err != nil {
		err := os.Mkdir(fmt.Sprintf("%spublic/media", dir), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}
	handleInterupt()
	router.Static("/public", "public")
	defer cleanUpDb()
	testSignUp(t, http.StatusCreated, "password", "abrahamakerele38@gmail.com", models.Manager)
	testSignUp(t, http.StatusCreated, "password", "abraham38@gmail.com", models.Landlord)
	testSignUp(t, http.StatusCreated, "password", "abrahamak38@gmail.com", models.Tenant)
	testSignUp(t, http.StatusCreated, "password", "niyi@gmail.com", models.Vendor)
	testSignIn(t, http.StatusOK, "password", "abrahamakerele38@gmail.com")
	testGetProfile(t, http.StatusOK)
	testChangePassword(t, http.StatusOK, "abrahamakerele38@gmail.com", "password", "newpassword")
	testSignIn(t, http.StatusBadRequest, "password", "abrahamakerele38@gmail.com")
	testSignIn(t, http.StatusOK, "newpassword", "abrahamakerele38@gmail.com")
	testResetPassword(t, http.StatusOK, "abrahamakerele38@gmail.com", "web")
	testChangePasswordByToken(t, http.StatusOK, "abrahamakerele38@gmail.com", "newpassword", "MTExMTEx")
	testResetPassword(t, http.StatusOK, "abrahamakerele38@gmail.com", "mobile")
	testChangePasswordByToken(t, http.StatusOK, "abrahamakerele38@gmail.com", "newpassword", "111111")
	testChangeUserProfile(t, http.StatusOK)
	testUpdateProfile(t, http.StatusOK)
	testCreateProperty(t, http.StatusCreated)
	testUpdateProperty(t, http.StatusOK)
	testAddLandlord(t, http.StatusOK, "09078918596", "mab", "abraham38@gmail.com")
	testAddLandlord(t, http.StatusOK, "09078918596", "mab", "newuser1@gmail.com")
	testListLandLord(t, http.StatusOK)
	testRemoveLandlord(t, http.StatusOK)
	testAddTenant(t, http.StatusOK, "09088918596", "mab inc", "abrahamak38@gmail.com")
	testAddTenant(t, http.StatusOK, "09088918596", "mab inc", "newuser2@gmail.com")
	testListTenant(t, http.StatusOK)
	testRemoveTenant(t, http.StatusOK)
	testRemoveAttachment(t, http.StatusOK, "images")
	testRemoveAttachment(t, http.StatusOK, "documents")
	testAddInspection(t, http.StatusCreated)
	testUpdateInspection(t, http.StatusOK)
	testRemoveInspection(t, http.StatusOK)
	testAddComplaints(t, http.StatusCreated)
	testAddComplaints(t, http.StatusCreated)
	testListComplaints(t, http.StatusOK)
	testUpdateComplaints(t, http.StatusOK)
	testListProperty(t, http.StatusOK)
	testListInspection(t, http.StatusOK)
	testUploadForm(t, http.StatusOK)
}

func TestUtils(t *testing.T) {
	if utils.CheckDateFormat("") != false {
		t.Fatal("Should have been false")
	}
	if utils.CheckDateFormat("9/11/2020") != true {
		t.Fatal("Should have been false")
	}
	if utils.CheckDateFormat("09/-1/020") != true {
		t.Fatal("Should have been false")
	}
}
func cleanUpDb() {
	os.Setenv("CLEAR","")
	if os.Getenv("CLEAR") == "CLEAR" {
		client := database.GetMongoDB().GetClient()
		log.Print(client.Database(database.DbName).Drop(context.TODO()))
		log.Println(os.RemoveAll(fmt.Sprintf("%spublic/media/", os.Getenv("ROOTDIR"))))
	}
}
