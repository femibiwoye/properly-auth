package models

type CreateProperty struct {
	Name    string
	Type    string
	Address string
}

type UpdatePropertyModel struct {
	Name    string
	Type    string
	Address string
	ID      string
}

type AddLandlord struct {
	UserID     string
	PropertyID string
}

func (a *AddLandlord) GetUserID() string {
	return a.UserID
}

func (a *AddLandlord) GetPropertyID() string {
	return a.PropertyID
}

type ListType struct {
	PropertyID string
}

type RemoveAttachmentModel struct {
	PropertyID     string
	AttachmentName string
	AttachmentType string
}

type ScheduleInspectionModel struct {
	PropertyID     string
	AttachmentName string
	AttachmentType string
}
