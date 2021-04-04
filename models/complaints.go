package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	//ComplaintsCollectionName holds the collection on mongodb where Complaints are being stored
	ComplaintsCollectionName = "Complaints"
)

const (
	Pending      = "Pending"
	Acknowledged = "Acknowledged"
	Resolved     = "Resolved"
)

//Complaints decribes a complaints  on properly
type Complaints struct {
	ID         string `json:"id"`
	CreatedAt  int64  `json:"created_at"`
	Text       string `json:"text"`
	CreatedBy  string `json:"created_by"`
	PropertyId string `json:"property_id"`
	Status     string `json:"status"`
}

func (c *Complaints) getID() string {
	return c.ID
}

func (c *Complaints) setID(id string) {
	c.ID = id
}

func (c *Complaints) getCreatedAt() int64 {
	return c.CreatedAt
}

func (c *Complaints) setCreatedAt(at int64) {
	c.CreatedAt = at
}

func ToComplaintsFromM(mongoM bson.M) (*Complaints, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	i := &Complaints{}
	err = bson.Unmarshal(uB, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

type ComplaintsModel struct {
	PropertyID string
	Text       string
	Date       int64
}

func (cm *ComplaintsModel) GetUserID() string {
	return ""
}

func (cm *ComplaintsModel) GetPropertyID() string {
	return cm.PropertyID
}

type UpdateComplaintsModel struct {
	PropertyID   string
	ComplaintsID string
	Text         string
	Date         int64
	Status       string
}
