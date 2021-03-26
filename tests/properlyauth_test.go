package test

import (
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
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

func initVar(t *testing.T) {

}
func TestScoodent(t *testing.T) {
	initVar(t)
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
	fmt.Println(dir, "ilnhdfkbhifdhbjkdfghdf")
	os.Setenv("ROOTDIR", dir)
	_, err = os.Stat(fmt.Sprintf("%spublic/media", dir))
	if err != nil {
		err := os.Mkdir(fmt.Sprintf("%spublic/media", dir), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}
	fmt.Println(err, "ilnhdfkbhifdhbjkdfghdf na ment")
	handleInterupt()
	router.Static("/public", "public")
	defer cleanUpDb()
	testSignUp(t, http.StatusCreated, "password", "abrahamakerele38@gmail.com")
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
}

func cleanUpDb() {
	// client := database.GetMongoDB().GetClient()
	// log.Print(client.Database(database.DbName).Drop(context.TODO()))
}
