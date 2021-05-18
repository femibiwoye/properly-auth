package chats

import (
	"encoding/json"
	"fmt"
	"log"
	"properlyauth/controllers"
	"properlyauth/models"
	"properlyauth/utils"
	"time"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"net/http"
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
		server.JoinRoom("/", s.ID(), s)
		err := models.Insert(
			&models.ChatSession{SessionID: s.ID(), CreatedAt: time.Now().Unix()},
			models.ChatSessionCollectionName)
		return err
	})

	server.OnEvent("/", "message", func(s socketio.Conn, msg string) string {
		lc := LiveChat{}
		if err := json.Unmarshal([]byte(msg), &lc); err == nil {
			res, err := utils.DecodeJWTToken(lc.Token)
			if err != nil {
				controllers.SendNotification("Couldn't parse token", "")
				return msg
			}
			chatSessionUser1, err := models.GetChatSession("sessionid", s.ID())
			if err != nil {
				log.Println(err.Error(), s.ID())
				controllers.SendNotification("Error fetching user chat session", chatSessionUser1.UserID)
				return msg
			}
			chatSessionUser1.UserID = res["user_id"]
			models.UpdateData(chatSessionUser1, models.ChatSessionCollectionName)
			ticker := time.NewTicker(500 * time.Millisecond)
			count := 0

			chatSessionUser2, err := models.GetChatSession("sessionid", lc.To)
			if err != nil {
			waiter:
				for {
					count++
					select {
					case t := <-ticker.C:
						chatSessionUser2, err = models.GetChatSession("sessionid", lc.To)
						if err == nil {
							break waiter
						} else {
							if count > 100 {
								//save the chat
								goto saveChat
							}
							log.Println("waiting for user ", t, lc.To)
						}

					}
				}
			}

			//TODO (check if the user is already in the room)
			server.JoinRoom("/", chatSessionUser2.SessionID, s)
			server.ForEach("/", chatSessionUser2.SessionID, func(c socketio.Conn) {
				buf, err := json.Marshal(lc)
				if err != nil {
					controllers.SendNotification("Error sending message", chatSessionUser1.UserID)
					return
				}
				c.Emit(fmt.Sprintf("%s", buf))
				controllers.SendNotification(lc.Text, chatSessionUser1.UserID)
			})

		saveChat:
			chat := models.Chats{
				CreatedAt: lc.CreatedAt,
				Medias:    lc.Medias,
				Text:      lc.Text,
				SentBy:    res["user_id"],
			}

			err = models.Insert(&chat, models.ChatCollectionName)
			if err != nil {
				controllers.SendNotification("Error storing  message", chatSessionUser1.UserID)
				return msg
			}

		} else {
			controllers.SendNotification("Invalid message sent", "")
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

// ListChatsBetween godoc
// @Summary list chat between users
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ListChatRequestModel true "requestdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /v1/list/chats/ [post]
// @Security ApiKeyAuth
func ListChatsBetween(c *gin.Context) {
	data := models.ListChatRequestModel{}
	user, _, ok := controllers.CheckUser(c, false)
	if !ok {
		return
	}
	c.ShouldBindJSON(&data)
	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}
	if len(errorResponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid fetch details"), errorResponse)
		return
	}
	chats, err := models.FetchDocByCriterionMultiple("sentby", models.ChatCollectionName, []string{data.OtherUserId, user.ID})
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of chats  between user"), chats)
}
