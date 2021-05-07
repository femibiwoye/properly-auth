package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	//InvitesCollectionName holds the collection on mongodb where invites are being stored
	InvitesCollectionName = "Invites"
)

//Invite decribes a invitation  on properly
type Invite struct {
	ID             string                 `json:"id"`
	CreatedAt      int64                  `json:"created_at"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Email          string                 `json:"email"`
	Phone          string                 `json:"phone"`
	AdditionalData map[string]interface{} `json:"additionaldata"`
	CreatedBy      string                 `json:"created_by"`
}

func (i *Invite) getID() string {
	return i.ID
}

func (i *Invite) setID(id string) {
	i.ID = id
}

func (i *Invite) getCreatedAt() int64 {
	return i.CreatedAt
}

func (i *Invite) setCreatedAt(at int64) {
	i.CreatedAt = at
}

func ToInviteFromM(mongoM bson.M) (*Invite, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	i := &Invite{}
	err = bson.Unmarshal(uB, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func GeInvite(criteria, value string) (*Invite, error) {
	m, err := FetchDocByCriterion(criteria, value, ComplaintsCollectionName)
	if err != nil {
		return nil, err
	}
	return ToInviteFromM(m)
}
