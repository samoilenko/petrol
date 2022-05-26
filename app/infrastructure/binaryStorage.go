package infrastructure

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var byteOrder = binary.LittleEndian

type BinaryStorage struct {
	handler io.ReadWriter
	logger  ILogger
}

func (b *BinaryStorage) Read(data interface{}) error {
	var err error
	//if reflect.ValueOf(data).Type().Kind() != reflect.Ptr {
	//	return fmt.Errorf("expected to got a pointer but got %s ", data)
	//}

	elem := reflect.TypeOf(data).Elem()
	elemType := elem.Kind()

	switch elemType {
	case reflect.Float32:
		err = b.readBytes(data)
		if err != nil {
			return err
		}
	case reflect.Int:
		err = b.readBytes(data)
		if err != nil {
			return err
		}
	case reflect.Int32:
		err = b.readBytes(data)
		if err != nil {
			return err
		}
	case reflect.String:
		s, err := b.readString()
		reflect.ValueOf(data).Elem().SetString(s)
		if err != nil {
			return err
		}
	case reflect.Struct:
		v := reflect.ValueOf(data)
		for i := 0; i < elem.NumField(); i += 1 {
			f := v.Elem().Field(i)
			if f.CanSet() != true || f.IsValid() == false {
				fmt.Println("CAN_NOT_SET")
				continue
			}

			if f.Type().Kind() == reflect.String {
				var s string
				s, err = b.readString()
				if err != nil {
					return err
				}
				f.SetString(s)
			}
			if f.Type().Kind() == reflect.Int32 {
				var j int32
				err = b.readBytes(&j)
				if err != nil {
					return err
				}
				f.SetInt(int64(j))
			}
			if f.Type().Kind() == reflect.Float32 {
				var j float32
				err = b.readBytes(&j)
				if err != nil {
					return err
				}
				f.SetFloat(float64(j))
			}

			if f.Type().Kind() == reflect.Struct {
				b.Read(f)
			}
			if f.Type().Kind() == reflect.Ptr {
				a := v.Elem().Field(i).Elem()
				fmt.Println(a)
				b.Read(v.Elem().Field(i).Interface())
			}
		}
	}

	return nil
}

func (b *BinaryStorage) readBytes(data interface{}) error {
	err := binary.Read(b.handler, byteOrder, data)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func (b *BinaryStorage) WriteString(s string) error {
	if err := b.Write(int32(len(s))); err != nil {
		return err
	}
	return b.Write([]byte(s))
}

func (b *BinaryStorage) readString() (string, error) {
	var length int32
	if err := b.readBytes(&length); err != nil {
		return "", err
	}
	bytes := make([]byte, length)
	if err := b.readBytes(&bytes); err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (b *BinaryStorage) Write(x interface{}) error {
	return binary.Write(b.handler, byteOrder, x)
}

func (b *BinaryStorage) WriteInt(x int32) error {
	return binary.Write(b.handler, byteOrder, x)
}

func NewBinaryStorage(logger ILogger, handler io.ReadWriter) *BinaryStorage {
	return &BinaryStorage{
		logger:  logger,
		handler: handler,
	}
}
