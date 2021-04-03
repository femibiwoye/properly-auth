package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	//UserCollectionName holds the collection on mongodb where user detailas are being stored
	InspectionCollectionaName = "Inspection"
)

//Inspection decribes a property inspection on properly
type Inspection struct {
	ID         string `json:"id"`
	CreatedAt  int64  `json:"created_at"`
	Text       string `json:"text"`
	DueTime    int64  `json:"duetime"`
	CreatedBy  string `json:"created_by"`
	PropertyId string `json:"property_id"`
}

func (i *Inspection) getID() string {
	return i.ID
}

func (i *Inspection) setID(id string) {
	i.ID = id
}

func (i *Inspection) getCreatedAt() int64 {
	return i.CreatedAt
}

func (i *Inspection) setCreatedAt(at int64) {
	i.CreatedAt = at
}

func ToInspectionFromM(mongoM bson.M) (*Inspection, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	i := &Inspection{}
	err = bson.Unmarshal(uB, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}
