package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	//ChatCollectionName holds chat record for various user
	NotificationCollectionName = "Notification"
)

//Complaints decribes a complaints  on properly
type Notification struct {
	ID         string   `json:"id"`
	CreatedAt  int64    `json:"created_at"`
	Text       string   `json:"text"`
	ReceivedBy string   `json:"receiveby"`
	Medias     []string `json:"medias"`
}

func (n *Notification) getID() string {
	return n.ID
}

func (n *Notification) setID(id string) {
	n.ID = id
}

func (n *Notification) getCreatedAt() int64 {
	return n.CreatedAt
}

func (n *Notification) setCreatedAt(at int64) {
	n.CreatedAt = at
}

func ToNotificationFromM(mongoM bson.M) (*Notification, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	i := &Notification{}
	err = bson.Unmarshal(uB, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func GetNotification(criteria, value string) (*Notification, error) {
	m, err := FetchDocByCriterion(criteria, value, ChatCollectionName)
	if err != nil {
		return nil, err
	}
	return ToNotificationFromM(m)
}
