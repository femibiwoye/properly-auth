package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
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
		log.Println("Connected", h.Id())
		time.Sleep(500 * time.Millisecond)
		for _, msg := range []string{
			randomStringMessage(5),
			randomStringMessage(5),
			randomStringMessage(5),
			randomStringMessage(5),
		} {

			lc := chats.LiveChat{To: h.Id(), Token: token}
			lc.CreatedAt = time.Now().Unix()
			lc.Text = msg
			buf, err := json.Marshal(lc)
			if err != nil {
				t.Fatalf(err.Error())
			}
			command := string(buf)
			c.Emit("message", command)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

}

func randomStringMessage(size int) string {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte((rand.Int() % 26) + 64)
	}
	return fmt.Sprintf("%s", buf)
}

func testListChat(t *testing.T, ExpectedCode int) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/list/chats/?platform=mobile", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens[0]))
	user2Res, err := utils.DecodeJWTToken(tokens[1])
	if err != nil {
		t.Fatal(user2Res, err)
	}
	data := make(map[string]interface{})
	data["OtherUserId"] = user2Res["user_id"]

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
	fmt.Println(string(responseText))
}
