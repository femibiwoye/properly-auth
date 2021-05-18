package chats

import (
	"encoding/json"
	"fmt"
	"log"
	"properlyauth/models"
	"properlyauth/utils"
	"time"

	socketio "github.com/googollee/go-socket.io"
)

type LiveChat struct {
	Token     string // user jwt token to authenticate the user
	To        string // the session ID of the user the message is for
	Text      string // text
	Medias    []string
	CreatedAt int64
}

// Event on websocket
var (
	StartChat = "startchat"
	CloseChat = "closechat"
	Message   = "message"
)

func CreateChatServer() *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		utils.PrintSomeThing(s.URL().RawQuery)
		server.JoinRoom("/", s.ID(), s)
		err := models.Insert(
			&models.ChatSession{SessionID: s.ID(), CreatedAt: time.Now().Unix()},
			models.ChatSessionCollectionName)
		log.Println(err)
		return err
	})

	server.OnEvent("/", "message", func(s socketio.Conn, msg string) string {
		lc := LiveChat{}
		if err := json.Unmarshal([]byte(msg), &lc); err == nil {
			res, err := utils.DecodeJWTToken(lc.Token)
			if err != nil {
				sendNotification("Couldn't parse token")
				return msg
			}

			chatSessionUser2, err := models.GetChatSession("sessionid", lc.To)
			if err != nil {
				//send an error message or Something
				sendNotification("Error fetching user chat session")
				return msg
			}

			//TODO (check if the user is already in the room)
			server.JoinRoom("/", chatSessionUser2.SessionID, s)
			server.ForEach("/", chatSessionUser2.SessionID, func(c socketio.Conn) {
				buf, err := json.Marshal(lc)
				if err != nil {
					//send some error notification here
					sendNotification("Error sending message")
					return
				}
				c.Emit(fmt.Sprintf("%s", buf))
				sendNotification(lc.Text)
			})

			chat := models.Chats{
				CreatedAt:  lc.CreatedAt,
				Medias:     lc.Medias,
				Text:       lc.Text,
				ReceivedBy: lc.To,
				SentBy:     res["user_id"],
			}

			err = models.Insert(&chat, models.ChatCollectionName)
			if err != nil {
				sendNotification("Error storing  message")
				return msg
			}

		} else {
			sendNotification("Invalid message sent")
			return msg
		}

		return msg
	})

	server.OnDisconnect("/", func(s socketio.Conn, msg string) {
		server.LeaveRoom("/", s.ID(), s)
		cs, err := models.GetChatSession("sessionid", s.ID())
		if err != nil {
			return
		}
		err = models.Delete(cs, models.ChatSessionCollectionName)
		if err != nil {
			return
		}
	})

	return server
}

func sendNotification(text string) {
	notification := models.Notification{
		Text: text,
	}
	log.Println(
		fmt.Sprintf("Inserting notification %s", notification.Text),
		models.Insert(&notification, models.NotificationCollectionName),
	)
}
