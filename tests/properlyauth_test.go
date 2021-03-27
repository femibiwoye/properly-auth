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

func TestScoodent(t *testing.T) {

	os.Setenv("HOST", "127.0.0.1:8080")
	os.Setenv("TESTING", "TESTING")
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
	testUploadPost(t, http.StatusOK)
	testCreateProperty(t, http.StatusCreated)
	testUpdateProperty(t, http.StatusOK)
	testAddLandlord(t, http.StatusOK)
	testRemoveLandlord(t, http.StatusOK)
	testAddTenant(t, http.StatusOK)
	testRemoveTenant(t, http.StatusOK)
}

func cleanUpDb() {
	if os.Getenv("CLEAR") == "CLEAR" {
		client := database.GetMongoDB().GetClient()
		log.Print(client.Database(database.DbName).Drop(context.TODO()))
		log.Println(os.RemoveAll(fmt.Sprintf("%spublic/media/", os.Getenv("ROOTDIR"))))
	}
}
