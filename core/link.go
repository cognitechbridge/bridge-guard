package core

type LinkData struct {
	ObjectId string `json:"objectId"`
	Size     int64  `json:"size"`
}

type Link struct {
	Path string
	Data LinkData
}

func (l *Link) Id() string {
	return l.Data.ObjectId
}
