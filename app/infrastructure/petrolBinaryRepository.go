package infrastructure

import "fmt"

type PetrolBinaryRepository struct {
	storage IBinaryStorage
}

func (pb *PetrolBinaryRepository) ReadAll() (map[string]*PetrolStationInfo, error) {
	var rowCount int32
	if err := pb.storage.Read(&rowCount); err != nil {
		return nil, err
	}

	res := make(map[string]*PetrolStationInfo, rowCount)
	var i int32
	for i = 0; i < rowCount; i++ {
		info := &PetrolStationInfo{Coordinates: &Coordinates{}}
		if err := pb.storage.Read(info); err != nil {
			return nil, err
		}

		res[info.Id] = info
	}

	return res, nil
}

func (pb *PetrolBinaryRepository) SaveAll(petrolStations map[string]*PetrolStationInfo) error {
	if err := pb.storage.WriteInt(int32(len(petrolStations))); err != nil {
		return err
	}

	for _, petrolStation := range petrolStations {
		if err := pb.savePetrolStations(petrolStation); err != nil {
			return err
		}
	}
	fmt.Println(int32(len(petrolStations)))
	return nil
}

func (pb *PetrolBinaryRepository) savePetrolStations(petrolInfo *PetrolStationInfo) error {
	for _, i := range []string{petrolInfo.Id, petrolInfo.PetrolType, petrolInfo.State, petrolInfo.Address} {
		if err := pb.storage.WriteString(i); err != nil {
			return err
		}
	}

	for _, i := range []float32{petrolInfo.Coordinates.Lat, petrolInfo.Coordinates.Lon} {
		if err := pb.storage.Write(i); err != nil {
			return err
		}
	}
	return nil
}

func NewPetrolBinaryRepository(storage IBinaryStorage) *PetrolBinaryRepository {
	return &PetrolBinaryRepository{
		storage: storage,
	}
}
