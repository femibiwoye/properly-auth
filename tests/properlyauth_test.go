package test

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"properlyauth/database"
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

func TestScoodent(t *testing.T) {
	os.Setenv("HOST", "127.0.0.1:8080")
	os.Setenv("TESTING", "TESTING")
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	handleInterupt()
	router.Static("/public", "public")
	defer cleanUpDb()
	testSignUp(t, http.StatusCreated, "password", "abrahamakerele38@gmail.com")
	testSignIn(t, http.StatusOK, "password", "abrahamakerele38@gmail.com")
	testGeneratePumc(t, http.StatusOK)
	testGetProfile(t, http.StatusOK)
	testChangePassword(t, http.StatusOK, "abrahamakerele38@gmail.com", "password", "newpassword")
	testSignIn(t, http.StatusBadRequest, "password", "abrahamakerele38@gmail.com")
	testSignIn(t, http.StatusOK, "newpassword", "abrahamakerele38@gmail.com")
	testResetPassword(t, http.StatusOK, "abrahamakerele38@gmail.com", "web")
	testChangePasswordByToken(t, http.StatusOK, "abrahamakerele38@gmail.com", "newpassword", "MTExMTEx")
	testResetPassword(t, http.StatusOK, "abrahamakerele38@gmail.com", "mobile")
	testChangePasswordByToken(t, http.StatusOK, "abrahamakerele38@gmail.com", "newpassword", "111111")
}

func cleanUpDb() {
	client := database.GetMongoDB().GetClient()
	log.Print(client.Database(database.DbName).Drop(context.TODO()))
}
