package models

type InspectionModel struct {
	userID     string
	PropertyID string
	Text       string
	Date       int64
}

func (im *InspectionModel) GetUserID() string {
	return im.userID
}

func (im *InspectionModel) GetPropertyID() string {
	return im.PropertyID
}

type UpdateInspectionModel struct {
	userID       string
	propertyID   string
	InspectionID string
	Text         string
	Date         int64
}

func (uim *UpdateInspectionModel) GetUserID() string {
	return uim.userID
}

func (uim *UpdateInspectionModel) GetPropertyID() string {
	return uim.propertyID
}

type InspectionDeleteModel struct {
	InspectionID string
}
