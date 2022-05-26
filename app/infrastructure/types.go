package infrastructure

type IHttpClient interface {
	Get(url string) ([]byte, error)
}

type IUnitOfWork interface {
	Persist()
	Save()
}

type Irepository interface {
}

type IBinaryStorage interface {
	Write(x interface{}) error
	WriteInt(x int32) error
	WriteString(s string) error
	Read(data interface{}) error
}

type ILogger interface {
	Error(args ...interface{})
	Info(args ...interface{})
}

type Coordinates struct {
	Lat float32
	Lon float32
}

type PetrolStationInfo struct {
	Id          string
	Address     string
	PetrolType  string
	State       string
	Coordinates *Coordinates
}
