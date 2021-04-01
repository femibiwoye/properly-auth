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

type ListType struct {
	PropertyID string
}

type RemoveAttachmentModel struct {
	PropertyID     string
	AttachmentName string
	AttachmentType string
}
