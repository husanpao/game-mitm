package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

// MessagePack format constants
const (
	// Nil
	FormatNil = 0xc0

	// Boolean
	FormatFalse = 0xc2
	FormatTrue  = 0xc3

	// Binary
	FormatBin8  = 0xc4
	FormatBin16 = 0xc5
	FormatBin32 = 0xc6

	// Extension
	FormatExt8     = 0xc7
	FormatExt16    = 0xc8
	FormatExt32    = 0xc9
	FormatFixExt1  = 0xd4
	FormatFixExt2  = 0xd5
	FormatFixExt4  = 0xd6
	FormatFixExt8  = 0xd7
	FormatFixExt16 = 0xd8

	// Float
	FormatFloat32 = 0xca
	FormatFloat64 = 0xcb

	// Unsigned Integer
	FormatUint8  = 0xcc
	FormatUint16 = 0xcd
	FormatUint32 = 0xce
	FormatUint64 = 0xcf

	// Signed Integer
	FormatInt8  = 0xd0
	FormatInt16 = 0xd1
	FormatInt32 = 0xd2
	FormatInt64 = 0xd3

	// String
	FormatStr8  = 0xd9
	FormatStr16 = 0xda
	FormatStr32 = 0xdb

	// Array
	FormatArray16 = 0xdc
	FormatArray32 = 0xdd

	// Map
	FormatMap16 = 0xde
	FormatMap32 = 0xdf
)

// Decoder 结构体用于解码MessagePack数据
type Decoder struct {
	data []byte
	pos  int
}

// NewDecoder 创建一个新的解码器
func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data: data,
		pos:  0,
	}
}

// UnpackMessagePack 解码MessagePack格式的字节数据
// 参数: data - 要解码的字节数组
// 返回: 解码后的数据和可能的错误
func UnpackMessagePack(data []byte) (interface{}, error) {
	decoder := NewDecoder(data)
	return decoder.Decode()
}

// UnpackFPNNMessage 解码FPNN协议包装的MessagePack数据
// 参数: data - 包含FPNN协议头的完整字节数组
// 返回: 解码后的MessagePack数据和可能的错误
func UnpackFPNNMessage(data []byte) (interface{}, error) {
	// 跳过FPNN协议头(通常是28字节)
	fpnnHeaderSize := 28
	if len(data) <= fpnnHeaderSize {
		return nil, fmt.Errorf("数据长度不足，无法跳过FPNN协议头")
	}

	msgpackData := data[fpnnHeaderSize:]
	return UnpackMessagePack(msgpackData)
}

func (d *Decoder) readByte() (byte, error) {
	if d.pos >= len(d.data) {
		return 0, fmt.Errorf("unexpected end of data")
	}
	b := d.data[d.pos]
	d.pos++
	return b, nil
}

func (d *Decoder) readBytes(n int) ([]byte, error) {
	if d.pos+n > len(d.data) {
		return nil, fmt.Errorf("unexpected end of data")
	}
	result := d.data[d.pos : d.pos+n]
	d.pos += n
	return result, nil
}

func (d *Decoder) readUint8() (uint8, error) {
	b, err := d.readByte()
	return uint8(b), err
}

func (d *Decoder) readUint16() (uint16, error) {
	bytes, err := d.readBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(bytes), nil
}

func (d *Decoder) readUint32() (uint32, error) {
	bytes, err := d.readBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(bytes), nil
}

func (d *Decoder) readUint64() (uint64, error) {
	bytes, err := d.readBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(bytes), nil
}

func (d *Decoder) readInt8() (int8, error) {
	b, err := d.readByte()
	return int8(b), err
}

func (d *Decoder) readInt16() (int16, error) {
	bytes, err := d.readBytes(2)
	if err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(bytes)), nil
}

func (d *Decoder) readInt32() (int32, error) {
	bytes, err := d.readBytes(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(bytes)), nil
}

func (d *Decoder) readInt64() (int64, error) {
	bytes, err := d.readBytes(8)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(bytes)), nil
}

func (d *Decoder) readFloat32() (float32, error) {
	bytes, err := d.readBytes(4)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.BigEndian.Uint32(bytes)), nil
}

func (d *Decoder) readFloat64() (float64, error) {
	bytes, err := d.readBytes(8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(bytes)), nil
}

// Decode 解码MessagePack数据
func (d *Decoder) Decode() (interface{}, error) {
	code, err := d.readByte()
	if err != nil {
		return nil, err
	}

	// Positive fixint
	if code <= 0x7f {
		return int(code), nil
	}

	// Fixmap
	if code >= 0x80 && code <= 0x8f {
		return d.decodeFixMap(int(code & 0x0f))
	}

	// Fixarray
	if code >= 0x90 && code <= 0x9f {
		return d.decodeFixArray(int(code & 0x0f))
	}

	// Fixstr
	if code >= 0xa0 && code <= 0xbf {
		return d.decodeFixStr(int(code & 0x1f))
	}

	// Negative fixint
	if code >= 0xe0 {
		return int(int8(code)), nil
	}

	switch code {
	case FormatNil:
		return nil, nil
	case FormatFalse:
		return false, nil
	case FormatTrue:
		return true, nil
	case FormatBin8:
		return d.decodeBin8()
	case FormatBin16:
		return d.decodeBin16()
	case FormatBin32:
		return d.decodeBin32()
	case FormatFloat32:
		return d.readFloat32()
	case FormatFloat64:
		return d.readFloat64()
	case FormatUint8:
		return d.readUint8()
	case FormatUint16:
		return d.readUint16()
	case FormatUint32:
		return d.readUint32()
	case FormatUint64:
		return d.readUint64()
	case FormatInt8:
		return d.readInt8()
	case FormatInt16:
		return d.readInt16()
	case FormatInt32:
		return d.readInt32()
	case FormatInt64:
		return d.readInt64()
	case FormatStr8:
		return d.decodeStr8()
	case FormatStr16:
		return d.decodeStr16()
	case FormatStr32:
		return d.decodeStr32()
	case FormatArray16:
		return d.decodeArray16()
	case FormatArray32:
		return d.decodeArray32()
	case FormatMap16:
		return d.decodeMap16()
	case FormatMap32:
		return d.decodeMap32()
	default:
		return nil, fmt.Errorf("unsupported format code: 0x%02x", code)
	}
}

func (d *Decoder) decodeFixStr(length int) (string, error) {
	bytes, err := d.readBytes(length)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (d *Decoder) decodeStr8() (string, error) {
	length, err := d.readUint8()
	if err != nil {
		return "", err
	}
	bytes, err := d.readBytes(int(length))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (d *Decoder) decodeStr16() (string, error) {
	length, err := d.readUint16()
	if err != nil {
		return "", err
	}
	bytes, err := d.readBytes(int(length))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (d *Decoder) decodeStr32() (string, error) {
	length, err := d.readUint32()
	if err != nil {
		return "", err
	}
	bytes, err := d.readBytes(int(length))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (d *Decoder) decodeBin8() ([]byte, error) {
	length, err := d.readUint8()
	if err != nil {
		return nil, err
	}
	return d.readBytes(int(length))
}

func (d *Decoder) decodeBin16() ([]byte, error) {
	length, err := d.readUint16()
	if err != nil {
		return nil, err
	}
	return d.readBytes(int(length))
}

func (d *Decoder) decodeBin32() ([]byte, error) {
	length, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	return d.readBytes(int(length))
}

func (d *Decoder) decodeFixArray(length int) ([]interface{}, error) {
	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		value, err := d.Decode()
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

func (d *Decoder) decodeArray16() ([]interface{}, error) {
	length, err := d.readUint16()
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, length)
	for i := 0; i < int(length); i++ {
		value, err := d.Decode()
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

func (d *Decoder) decodeArray32() ([]interface{}, error) {
	length, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, length)
	for i := 0; i < int(length); i++ {
		value, err := d.Decode()
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

func (d *Decoder) decodeFixMap(length int) (map[interface{}]interface{}, error) {
	result := make(map[interface{}]interface{})
	for i := 0; i < length; i++ {
		key, err := d.Decode()
		if err != nil {
			return nil, err
		}
		value, err := d.Decode()
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func (d *Decoder) decodeMap16() (map[interface{}]interface{}, error) {
	length, err := d.readUint16()
	if err != nil {
		return nil, err
	}
	result := make(map[interface{}]interface{})
	for i := 0; i < int(length); i++ {
		key, err := d.Decode()
		if err != nil {
			return nil, err
		}
		value, err := d.Decode()
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func (d *Decoder) decodeMap32() (map[interface{}]interface{}, error) {
	length, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	result := make(map[interface{}]interface{})
	for i := 0; i < int(length); i++ {
		key, err := d.Decode()
		if err != nil {
			return nil, err
		}
		value, err := d.Decode()
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}
