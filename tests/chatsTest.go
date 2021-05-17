package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"properlyauth/controllers/chats"
	"properlyauth/routes"
	"properlyauth/utils"
	"testing"
	"time"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func testChat(t *testing.T) {
	router, chatServer := routes.Router()
	defer chatServer.Close()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	time.Sleep(time.Second * 1)
	user1Res, err := utils.DecodeJWTToken(tokens[0])
	if err != nil {
		t.Fatal(user1Res, err)
	}
	user2Res, err := utils.DecodeJWTToken(tokens[1])
	if err != nil {
		t.Fatal(user2Res, err)
	}

	sendMessages(t, user1Res["user_id"], tokens[0])
	sendMessages(t, user2Res["user_id"], tokens[1])

	time.Sleep(time.Second * 5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

}

func sendMessages(t *testing.T, to, token string) {
	c, err := gosocketio.Dial(
		gosocketio.GetUrl("127.0.0.1", 8080, false),
		transport.GetDefaultWebsocketTransport(),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		c.Join(h.Id())
		log.Println("Connected")
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, msg := range []string{
		randomStringMessage(5),
	} {

		lc := chats.LiveChat{To: to, Token: token}
		lc.CreatedAt = time.Now().Unix()
		lc.Text = msg
		buf, err := json.Marshal(lc)
		if err != nil {
			t.Fatalf(err.Error())
		}
		command := string(buf)
		c.Emit("message", command)
	}

}

func randomStringMessage(size int) string {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte(rand.Int() % 255)
	}
	return fmt.Sprintf("%s", buf)
}
