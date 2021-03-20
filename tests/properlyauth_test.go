package test

import (
	"context"
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
	handleInterupt()
	router.Static("/public", "public")
	defer cleanUpDb()
	testSignUp(t, http.StatusBadRequest, "+2349xhbsvhs078918596")
	testSignUp(t, http.StatusCreated, "+2349078918596")

}

func cleanUpDb() {
	client := database.GetMongoDB().GetClient()
	log.Print(client.Database(database.DbName).Drop(context.TODO()))
}
