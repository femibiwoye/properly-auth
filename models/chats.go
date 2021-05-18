package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	//ChatCollectionName holds chat record for various user
	ChatCollectionName        = "Chats"
	ChatSessionCollectionName = "ChatSession"
)

//Complaints decribes a complaints  on properly
type Chats struct {
	ID        string   `json:"id"`
	CreatedAt int64    `json:"created_at"`
	Text      string   `json:"text"`
	SentBy    string   `json:"sentby"`
	Medias    []string `json:"medias"`
}

func (c *Chats) getID() string {
	return c.ID
}

func (c *Chats) setID(id string) {
	c.ID = id
}

func (c *Chats) getCreatedAt() int64 {
	return c.CreatedAt
}

func (c *Chats) setCreatedAt(at int64) {
	c.CreatedAt = at
}

func ToChatsFromM(mongoM bson.M) (*Chats, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	i := &Chats{}
	err = bson.Unmarshal(uB, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func GetChats(criteria, value string) (*Chats, error) {
	m, err := FetchDocByCriterion(criteria, value, ChatCollectionName)
	if err != nil {
		return nil, err
	}
	return ToChatsFromM(m)
}

type ChatSession struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"createdat"`
	UserID    string `json:"userid"`
	SessionID string `json:"sessionid"`
}

func (c *ChatSession) getID() string {
	return c.ID
}

func (c *ChatSession) setID(id string) {
	c.ID = id
}

func (c *ChatSession) getCreatedAt() int64 {
	return c.CreatedAt
}

func (c *ChatSession) setCreatedAt(at int64) {
	c.CreatedAt = at
}

func ToChatSessionFromM(mongoM bson.M) (*ChatSession, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	i := &ChatSession{}
	err = bson.Unmarshal(uB, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func GetChatSession(criteria, value string) (*ChatSession, error) {
	m, err := FetchDocByCriterion(criteria, value, ChatSessionCollectionName)
	if err != nil {
		return nil, err
	}
	return ToChatSessionFromM(m)
}

type ListChatRequestModel struct {
	OtherUserId string
}
