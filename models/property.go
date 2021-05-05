package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	//PropertyCollectionName holds the collection for property
	PropertyCollectionName = "Property"
)

//Propertu decribes user property on properly
type Property struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Address   string            `json:"address"`
	Images    []string          `json:"images"`
	Documents []string          `json:"documents"`
	Forms     []string          `json:"forms"`
	Landlords map[string]string `json:"landlord"`
	Tenants   map[string]string `json:"tenants"`
	Vendors   map[string]string `json:"vendors"`
	CreatedAt int64             `json:"created_at"`
	CreatedBy string            `json:"created_by"`
}

func (p *Property) getID() string {
	return p.ID
}

func (p *Property) setID(id string) {
	p.ID = id
}

func (p *Property) getCreatedAt() int64 {
	return p.CreatedAt
}

func (p *Property) setCreatedAt(at int64) {
	p.CreatedAt = at
}

func ToPropertyFromM(mongoM bson.M) (*Property, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	p := &Property{}
	err = bson.Unmarshal(uB, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
