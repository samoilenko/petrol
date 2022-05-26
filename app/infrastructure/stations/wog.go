package stations

import (
	"encoding/json"
	"errors"
	"fmt"
	"petrol/infrastructure"
	"reflect"
	"strings"
	"time"
)

type WOG struct {
	httpClient         infrastructure.IHttpClient
	logger             infrastructure.ILogger
	url                string
	oldState           map[string]*infrastructure.PetrolStationInfo
	allowedPetrolTypes map[string]struct{}
}

func (o *WOG) Operate(res chan *infrastructure.PetrolStationInfo, delay time.Duration) {
	// scan -> filter(is changed) -> saveToOld -> return
	infrastructure.ExecutePipeline(
		infrastructure.Job(func(in, out chan interface{}) {
			for {
				o.scan(out)

				time.Sleep(delay)
			}
		}),
		//infrastructure.Job(o.getWithNewState),
		//infrastructure.Job(o.storeChanged),
		infrastructure.Job(func(in, out chan interface{}) {
			for data := range in {
				res <- data.(*infrastructure.PetrolStationInfo)
			}
		}),
	)
}

//
//func (o *WOG) storeChanged(in, out chan interface{}) {
//	for data := range in {
//		petrolInfo := data.(*infrastructure.PetrolStationInfo)
//		o.oldState[petrolInfo.PetrolType] = petrolInfo
//		out <- data
//	}
//}
//
//func (o *WOG) getWithNewState(in, out chan interface{}) {
//	for info := range in {
//		petrol := info.(*infrastructure.PetrolStationInfo)
//		_, exists := o.oldState[petrol.PetrolType]
//		if !exists || o.oldState[petrol.PetrolType].State != petrol.State {
//			out <- petrol
//		}
//	}
//}

func (o *WOG) scan(out chan interface{}) {
	data, err := o.httpClient.Get(o.url)
	if err != nil {
		o.logger.Error(err)
		return
	}
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	info := result["data"].(map[string]interface{})
	stationId, err := getInt(info["id"])

	if err != nil {
		o.logger.Error("can't get petrol station id", err)
		return
	}

	address, err := getString(info["name"])

	if err != nil {
		o.logger.Error("can't get address", err)
		return
	}

	petrolInfo, err := getPetrolInfo(info["workDescription"], o.allowedPetrolTypes)
	if err != nil {
		o.logger.Error("can't get petrol info", err)
		return
	}

	coordinates, err := getCoordinates(info["coordinates"])
	if err != nil {
		o.logger.Error("can't get coordinates", err)
		return
	}

	for petrolType, state := range petrolInfo {
		out <- &infrastructure.PetrolStationInfo{
			Id:          fmt.Sprintf("%d%s", stationId, petrolType),
			Address:     address,
			PetrolType:  petrolType,
			State:       state,
			Coordinates: coordinates,
		}
	}
}

func itemByKeyName(reflectedMap reflect.Value, name string) (reflect.Value, error) {
	mapKeys := reflectedMap.MapKeys()

	for _, key := range mapKeys {
		if name == key.String() {
			return reflectedMap.MapIndex(key), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("the key %s is absent", name)
}

func getCoordinates(data interface{}) (*infrastructure.Coordinates, error) {
	reflectedData := reflect.ValueOf(data)

	if reflectedData.Kind() != reflect.Map {
		return nil, errors.New("data is not struct")
	}

	coordinates := new(infrastructure.Coordinates)
	coordinatesElem := reflect.TypeOf(coordinates).Elem()
	coordinatesElemValue := reflect.ValueOf(coordinates).Elem()
	mapNames := map[string]string{"Lat": "latitude", "Lon": "longitude"}

	numFields := coordinatesElem.NumField()

	for i := 0; i < numFields; i++ {
		f := coordinatesElemValue.Field(i)

		if f.CanSet() != true || f.IsValid() == false {
			continue
		}

		mappedName, ok := mapNames[coordinatesElem.Field(i).Name]

		if !ok {
			return nil, fmt.Errorf("Can't find a mapped key for '%s'", coordinatesElem.Field(i).Name)
		}
		value, err := itemByKeyName(reflectedData, mappedName)

		if err != nil {
			return nil, err
		}

		if coordinatesElem.Field(i).Type.Kind() == reflect.Float32 {
			newValue, ok := value.Interface().(float64)
			if !ok {
				return nil, fmt.Errorf("the %s must be a float value", value.Interface())
			}

			f.SetFloat(newValue)
		}
	}

	return coordinates, nil
}

func getString(data interface{}) (string, error) {
	val := reflect.ValueOf(data)
	reflectType := val.Type()

	if reflectType.Kind() != reflect.String {
		return "", errors.New("data is not string")
	}

	return val.String(), nil
}

func getInt(data interface{}) (int32, error) {
	val := reflect.ValueOf(data)
	reflectType := val.Type()

	if reflectType.Kind() != reflect.Float64 {
		return 0, fmt.Errorf("data is %s, expected float64", reflectType.Kind().String())
	}

	return int32(val.Float()), nil
}

func getPetrolInfo(data interface{}, allowedPetrolTypes map[string]struct{}) (map[string]string, error) {
	infoAsString, err := getString(data)

	if err != nil {
		return nil, err
	}

	parts := strings.Split(infoAsString, "\n")

	var res = make(map[string]string)

	for _, part := range parts {
		petrolInfo := strings.SplitN(part, " ", 2)

		if _, ok := allowedPetrolTypes[petrolInfo[0]]; !ok { // != "М95" || petrolInfo[0] != "А95" {
			continue
		}
		res[petrolInfo[0]] = petrolInfo[1]
	}

	return res, nil
}

func NewWog(url string, allowedPetrolTypes map[string]struct{}, httpClient infrastructure.IHttpClient, logger infrastructure.ILogger) *WOG {
	return &WOG{
		url:                url,
		allowedPetrolTypes: allowedPetrolTypes,
		httpClient:         httpClient,
		logger:             logger,
		oldState:           make(map[string]*infrastructure.PetrolStationInfo),
	}
}
