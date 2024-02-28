package core

type ObjectService interface {
	Read(id string, buff []byte, ofst int64) (int, error)
	Write(id string, buff []byte, ofst int64) (int, error)
	Create(id string) error
	Move(oldId string, newId string) error
	Truncate(id string, size int64) error
	IsInQueue(id string) bool
	GetKeyIdByObjectId(id string) (string, error)
}
