package serialization

import (
	"encoding/binary"
	"math"
)

type AutoSerializeArray struct {
	bytes [][]byte
}
type AutoSerializeDic struct {
	dic map[string] []byte
}

// constructor
func NewAutoSerializedArray() *AutoSerializeArray {
	return new(AutoSerializeArray)
}

func NewAutoSerializedDic() *AutoSerializeDic {
	var instance = new(AutoSerializeDic)
	instance.dic = make(map[string] []byte)
	return instance
}

func (s *AutoSerializeArray)Count() int {
	return len(s.bytes)
}

func (s *AutoSerializeDic)Count() int {
	return len(s.dic)
}

// Array Append
func (s *AutoSerializeArray) SetSerializedBytes(bytes *[]byte) {
	if bytes == nil {
		return
	}
	bytesLen := len(*bytes)
	cursor := 0
	for cursor < bytesLen {
		bs := (*bytes)[cursor: cursor + 4]
		cursor += 4
		dataLen := int(binary.LittleEndian.Uint32(bs))
		s.bytes = append(s.bytes, (*bytes)[cursor: cursor + dataLen])
		cursor += dataLen
	}
}

func (s *AutoSerializeArray) AppendBytes(bytes *[]byte) {
	if bytes == nil {
		return
	}
	if len(*bytes) > 0 {
		s.bytes = append(s.bytes, *bytes)
	}
}

func (s *AutoSerializeArray) AppendInt(intger int32) {
	intBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(intBytes, uint32(intger))
	s.AppendBytes(&intBytes)
}

func (s *AutoSerializeArray) AppendFloat32(float float32) {
	floatBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(floatBytes, math.Float32bits(float))
	s.AppendBytes(&floatBytes)
}

func (s *AutoSerializeArray) AppendFloat64(float float64) {
	floatBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(floatBytes, math.Float64bits(float))
	s.AppendBytes(&floatBytes)
}

func (s *AutoSerializeArray) AppendString(str *string) {
	if str == nil {
		return
	}
	strBytes := []byte(*str)
	s.AppendBytes(&strBytes)
}

func (s *AutoSerializeArray) AppendArray(array *AutoSerializeArray){
	bytes := array.SerializedBytes()
	if bytes != nil && len(*bytes) > 0 {
		s.AppendBytes(bytes)
	}
}

func (s *AutoSerializeArray) AppendDic(dic *AutoSerializeDic) {
	bytes := dic.SerializedBytes()
	if bytes != nil && len(*bytes) > 0 {
		s.AppendBytes(bytes)
	}
}

// Array Accesor
func (s *AutoSerializeArray) BytesAtIndex(idx int) (*[]byte, bool) {
	if idx < len(s.bytes) {
		bs := s.bytes[idx]
		return &bs, true
	}
	return nil, false
}

func (s *AutoSerializeArray) IntegerAtIndex(idx int) (int32, bool) {
	bs, _ := s.BytesAtIndex(idx)
	if len(*bs) != 4 {
		return 0, false
	}
	integer := binary.LittleEndian.Uint32(*bs)
	return int32(integer), true
}

func (s *AutoSerializeArray) Float32AtIndex(idx int) (float32, bool) {
	bs, _ := s.BytesAtIndex(idx)
	if len(*bs) != 4 {
		return 0, false
	}
	integer := binary.LittleEndian.Uint32(*bs)
	return math.Float32frombits(integer), true
}

func (s *AutoSerializeArray) Float64AtIndex(idx int) (float64, bool) {
	bs, _ := s.BytesAtIndex(idx)
	if len(*bs) != 8 {
		return 0, false
	}
	integer := binary.LittleEndian.Uint64(*bs)
	return math.Float64frombits(integer), true
}

func (s *AutoSerializeArray) StringAtIndex(idx int) (*string, bool) {
	bs, _ := s.BytesAtIndex(idx)
	if bs != nil {
		str := string(*bs)
		return &str, true
	} else {
		return nil, false
	}
}

func (s *AutoSerializeArray) ArrayAtIndex(idx int) (*AutoSerializeArray, bool) {
	bs, _ := s.BytesAtIndex(idx)
	if bs != nil {
		array := NewAutoSerializedArray()
		array.SetSerializedBytes(bs)
		return array, true
	} else {
		return nil, false
	}
}

func (s *AutoSerializeArray) DicAtIndex(idx int) (*AutoSerializeDic, bool) {
	bs, _ := s.BytesAtIndex(idx)
	if bs != nil {
		dic := NewAutoSerializedDic()
		dic.SetSerializedBytes(bs)
		return dic, true
	} else {
		return nil, false
	}
}

// genearete Data
func (s *AutoSerializeArray) SerializedBytes() *[]byte {
	var bs []byte
	for _, v := range s.bytes {
		dataLenBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(dataLenBytes, uint32(len(v)))
		bs = append(bs, dataLenBytes...)
		bs = append(bs, v...)
	}
	return &bs
}


// SetObject
func (s *AutoSerializeDic) SetSerializedBytes(bytes *[]byte) {
	if bytes == nil {
		return
	}
	bytesLen := len(*bytes)
	cursor := 0
	for cursor < bytesLen {
		keyLen := int((*bytes)[cursor])
		cursor += 1
		bs := (*bytes)[cursor: cursor + 4]
		dataLen := int(binary.LittleEndian.Uint32(bs))
		cursor += 4
		keyBytes := (*bytes)[cursor: cursor + keyLen]
		cursor += keyLen
		dataBytes := (*bytes)[cursor: cursor + dataLen]
		cursor += dataLen
		s.dic[string(keyBytes)] = dataBytes
	}
}

func (s *AutoSerializeDic) SetBytes(bytes *[]byte, key string) {
	if bytes == nil {
		return
	}
	if len(*bytes) > 0 && len(key) > 0 {
		s.dic[key] = *bytes
	}
}

func (s *AutoSerializeDic) SetInt(intger int32, key string) {
	if len(key) > 0 {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(intger))
		s.SetBytes(&bs, key)
	}
}

func (s *AutoSerializeDic) SetFloat32(float float32, key string) {
	if len(key) > 0 {
		floatBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(floatBytes, math.Float32bits(float))
		s.SetBytes(&floatBytes, key)
	}
}

func (s *AutoSerializeDic) SetFloat64(float float64, key string) {
	if len(key) > 0 {
		floatBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(floatBytes, math.Float64bits(float))
		s.SetBytes(&floatBytes, key)
	}
}

func (s *AutoSerializeDic) SetString(str *string, key string) {
	if str == nil {
		return
	}
	if len(*str) > 0 && len(key) > 0 {
		bs := []byte(*str)
		s.SetBytes(&bs, key)
	}
}

func (s *AutoSerializeDic) SetArray(array *AutoSerializeArray, key string){
	bytes := array.SerializedBytes()
	if bytes != nil && len(*bytes) > 0 && len(key) > 0 {
		s.SetBytes(bytes, key)
	}
}

func (s *AutoSerializeDic) SetDic(dic *AutoSerializeDic, key string) {
	bytes := dic.SerializedBytes()
	if bytes != nil && len(*bytes) > 0 && len(key) > 0 {
		s.SetBytes(bytes, key)
	}
}

// Dic Accesor
func (s *AutoSerializeDic) BytesWithKey(key string) (*[]byte, bool) {
	if len(key) > 0 {
		bs := s.dic[key]
		if len(bs) > 0 {
			return &bs, true
		}
	}
	return nil, false
}

func (s *AutoSerializeDic) IntegerWithKey(key string) (int32, bool) {
	bs, _ := s.BytesWithKey(key)
	if len(*bs) != 4 {
		return 0, false
	}
	integer := binary.LittleEndian.Uint32(*bs)
	return int32(integer), true
}

func (s *AutoSerializeDic) Uint64WithKey(key string) (uint64, bool) {
	bs, _ := s.BytesWithKey(key)
	if len(*bs) != 8 {
		return 0, false
	}
	integer := binary.LittleEndian.Uint64(*bs)
	return uint64(integer), true
}

func (s *AutoSerializeDic) Float32WithKey(key string) (float32, bool) {
	bs, _ := s.BytesWithKey(key)
	if len(*bs) != 4 {
		return 0, false
	}
	integer := binary.LittleEndian.Uint32(*bs)
	return math.Float32frombits(integer), true
}

func (s *AutoSerializeDic) Float64WithKey(key string) (float64, bool) {
	bs, _ := s.BytesWithKey(key)
	if len(*bs) != 8 {
		return 0, false
	}
	integer := binary.LittleEndian.Uint64(*bs)
	return math.Float64frombits(integer), true
}

func (s *AutoSerializeDic) StringWithKey(key string) (*string, bool) {
	bs, _ := s.BytesWithKey(key)
	if bs != nil {
		str := string(*bs)
		return &str, true
	} else {
		return nil, false
	}
}

func (s *AutoSerializeDic) ArrayWithKey(key string) (*AutoSerializeArray, bool) {
	bs, _ := s.BytesWithKey(key)
	if bs != nil {
		array := NewAutoSerializedArray()
		array.SetSerializedBytes(bs)
		return array, true
	} else {
		return nil, false
	}
}

func (s *AutoSerializeDic) DicWithKey(key string) (*AutoSerializeDic, bool) {
	bs, _ := s.BytesWithKey(key)
	if bs != nil {
		dic := NewAutoSerializedDic()
		dic.SetSerializedBytes(bs)
		return dic, true
	} else {
		return nil, false
	}
}

func (s *AutoSerializeDic) Allkeys() ([]string) {
	var allKeys []string
	for k := range s.dic {
		allKeys = append(allKeys, k)
	}
	return allKeys
}

// genearete Data
func (s *AutoSerializeDic) SerializedBytes() *[]byte {
	var bs []byte
	for k, v := range s.dic {
		keylenByte := byte(len(k))
		datalenByte := make([]byte, 4)
		binary.LittleEndian.PutUint32(datalenByte, uint32(len(v)))
		bs = append(bs, keylenByte)
		bs = append(bs, datalenByte...)
		bs = append(bs, k...)
		bs = append(bs, v...)
	}
	return &bs
}