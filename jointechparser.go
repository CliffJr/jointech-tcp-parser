// Copyright 2023 Filip Kroča.
package jointechparser

import (
	"fmt"
	"github.com/CliffJr/b2n"
	"strconv"
)

// Decoded struct represent decoded E-Lock JointTech data structure with all ACL(Automated Container Lock) data as return from function Decode
type Decoded struct {
	ProtocolHeader        uint8
	ProtocolVersion       string
	IMEI                  string
	TerminalID            string
	Date                  string
	DeviceType            string
	DataType              string
	DataLength            string
	DirectionIndicator    string
	Mileage               int64
	BindVehicleID         string
	DeviceStatus          string
	BatteryLevel          uint8
	CellIdPositionCode    string
	GSMSignalQuality      uint8
	FenceAlarmID          uint8
	MNCHighByte           uint8
	ExpandedDeviceStatus  uint8
	ExpandedDeviceStatus2 uint8
	DataSerialNo          uint8
	Data                  []ACLData // Slice with ACL data
}

// ACLData represent one block of data
type ACLData struct {
	UtimeMs  uint64    // Utime is Time in mili seconds
	Utime    uint64    // Utime is Time in seconds
	Priority uint8     // JT does not provide this value
	Lat      int32     // Latitude (between 850000000 and -850000000), fit int32
	Lng      int32     // Longitude (between 1800000000 and -1800000000), fit int32
	Altitude int16     // JT does not provide this value
	Angle    uint16    // Direction in degrees from the JT docs In degrees
	VisSat   uint8     // The number of GPS satellites
	Speed    uint16    // Speed in km/h
	EventID  uint16    // JT does not provide this value
	Elements []Element // Slice containing parsed Elements
}

// Element represent one IO element, before storing in a db do a conversion to IO datatype (1B, 2B, 4B, 8B)
type Element struct {
	Length uint16 // Length of element, this should be uint16 because Codec 8 extended has 2Byte of IO len
	Name   uint16 // IO element ID
	Value  []byte // Value of the element represented by slice of bytes
}

// DeviceStates - various states and alarms of the device
type DeviceStates struct {
	Name        string
	Value       bool
	Description string
}

// Decode takes a pointer to a slice of bytes with raw data and return Decoded struct
func Decode(bs *[]byte) (Decoded, error) {
	decoded := Decoded{}
	var err error
	var nextByte int

	// check for minimum packet size
	if len(*bs) < 45 {
		return Decoded{}, fmt.Errorf("Minimum packet size is 45 Bytes, got %v", len(*bs))
	}

	// check for JT packet validity
	if (*bs)[0] != 0x24 && (*bs)[0] != 0x28 {
		return Decoded{}, fmt.Errorf("Probably not JT packet, trashed")
	}

	// determine bit number where start data, it can change because of IMEI length
	imeiLenX, err := b2n.ParseBs2Uint8(bs, 7)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}
	imeiLen := int(imeiLenX)

	if imeiLen != 15 && imeiLen != 16 {
		//log.Fatalf("Error when determining IMEI len want 15 or 16, got %v", imeiLen)
		return Decoded{}, fmt.Errorf("Error when determining IMEI len want 15 or 16, got %v", imeiLen)
	}

	// decode and validate IMEI
	decoded.IMEI, err = b2n.ParseIMEI(bs, 8, imeiLen)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// count start bit for data
	startByte := 0

	// decode Protocol Header
	decoded.ProtocolHeader = (*bs)[startByte]
	if decoded.ProtocolHeader != 0x24 && decoded.ProtocolHeader != 0x28 {
		return Decoded{}, fmt.Errorf("Invalid Protocol Header, want 0x24 or 0x28, get %v", decoded.ProtocolHeader)
	}

	// initialize nextByte counter
	nextByte = startByte

	// determine protocol header in packet
	decoded.ProtocolHeader, err = b2n.ParseBs2Uint8(bs, nextByte)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// increment nextByte counter
	nextByte++

	decoded.TerminalID, err = b2n.ParseBs2String(bs, nextByte, 10)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	nextByte = nextByte + 5

	// determine protocol version in packet
	parsedProtocol, err := b2n.ParseBs2String(bs, 6, 2)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert uint64 error, %v", err)
	}

	decoded.ProtocolVersion = parsedProtocol

	// increment nextByte counter
	nextByte++

	// determine device type in packet
	decoded.DeviceType, err = b2n.ParseBs2String(bs, nextByte, 2)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine data type in packet
	decoded.DataType, err = b2n.ParseBs2String(bs, nextByte, 2)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	nextByte++

	// determine data length in packet
	decoded.DataLength, err = b2n.ParseBs2String(bs, nextByte, 4)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	nextByte = nextByte + 2

	// determine date in packet
	decoded.Date, err = b2n.ParseBs2String(bs, nextByte, 6)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	nextByte = nextByte + 2

	// determine date in packet
	decoded.Date, err = b2n.ParseBs2String(bs, nextByte, 6)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine direction indicator in packet
	decoded.DirectionIndicator, err = b2n.ParseBs2String(bs, 20, 1)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine mileage in packet
	parseMileage, err := b2n.ParseBs2String(bs, 27, 8)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	parsedMileage, err := strconv.ParseInt(parseMileage, 10, 64)
	if err != nil {
		panic(err)
	}

	decoded.Mileage = parsedMileage

	// determine bind vehicle id in packet
	decoded.BindVehicleID, err = b2n.ParseBs2String(bs, 32, 8)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine device status in packet
	decoded.DeviceStatus, err = b2n.ParseBs2String(bs, 36, 4)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine battery level in packet
	decoded.BatteryLevel, err = b2n.ParseBs2Uint8(bs, 38)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine Cell Id Position Code in packet
	decoded.CellIdPositionCode, err = b2n.ParseBs2String(bs, 39, 8)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine GSM quality in packet
	decoded.GSMSignalQuality, err = b2n.ParseBs2Uint8(bs, 43)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine Fence Alarm ID in packet
	decoded.FenceAlarmID, err = b2n.ParseBs2Uint8(bs, 44)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine MNC High Byte  in packet
	decoded.MNCHighByte, err = b2n.ParseBs2Uint8(bs, 46)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine ExpandedDeviceStatus in packet
	decoded.ExpandedDeviceStatus, err = b2n.ParseBs2Uint8(bs, 45)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine ExpandedDeviceStatus2 in packet
	decoded.ExpandedDeviceStatus2, err = b2n.ParseBs2Uint8(bs, 47)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// make slice for decoded data
	decoded.Data = make([]ACLData, 0, len(decoded.Data))

	// go through data
	for i := 0; i < len(decoded.Data); i++ {

		decodedData := ACLData{}

		// time record in ms has 8 Bytes
		parsedTime, err := b2n.ParseBs2String(bs, 13, 6)

		// Convert string to uint64
		parsedTimeUint64, err := strconv.ParseUint(parsedTime, 10, 64)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert uint64 error, %v", err)
		}

		decodedData.UtimeMs = parsedTimeUint64
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error, %v", err)
		}

		decodedData.Utime = uint64(decodedData.UtimeMs / 1000)

		// parse priority will be nil because JT does not provide that value
		decodedData.Priority = 0

		// parse lat and validate GPS
		parsedLat, err := b2n.ParseBs2String(bs, 16, 8)

		// Convert string to uint32
		parsedLatInt32, err := strconv.ParseUint(parsedLat, 10, 64)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error, %v", err)
		}

		decodedData.Lat = int32(parsedLatInt32)

		if !(decodedData.Lat > -850000000 && decodedData.Lat < 850000000) {
			return Decoded{}, fmt.Errorf("Invalid Lat value, want lat > -850000000 AND lat < 850000000, got %v", decodedData.Lat)
		}

		// parse Lng and validate GPS
		parsedLng, err := b2n.ParseBs2String(bs, 20, 9)

		// Convert string to uint32
		parsedLngInt32, err := strconv.ParseUint(parsedLng, 10, 64)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error, %v", err)
		}

		decodedData.Lng = int32(parsedLngInt32)

		if !(decodedData.Lng > -1800000000 && decodedData.Lng < 1800000000) {
			return Decoded{}, fmt.Errorf("Invalid Lat value, want lat > -1800000000 AND lat < 1800000000, got %v", decodedData.Lng)
		}

		// JT does not provide the Altitude
		decodedData.Altitude = 0

		// parse Angle
		parsedAngle, err := b2n.ParseBs2String(bs, 26, 2)

		// Convert string to float32
		parsedLngFloat32, err := strconv.ParseUint(parsedAngle, 10, 32)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error, %v", err)
		}

		decodedData.Angle = uint16(parsedLngFloat32)

		if decodedData.Angle > 360 {
			return Decoded{}, fmt.Errorf("Invalid Angle value, want Angle <= 360, got %v", decodedData.Angle)
		}

		// parse num. of visible satellites VisSat
		parsedSatellites, err := b2n.ParseBs2String(bs, 31, 2)

		// Convert string to float32
		parsedSatsFloat32, err := strconv.ParseUint(parsedSatellites, 10, 32)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error, %v", err)
		}

		decodedData.VisSat = uint8(parsedSatsFloat32)

		// parse Speed
		parsedSpeed, err := b2n.ParseBs2String(bs, 6, 2)

		// Convert string to float32
		parsedSpeedFloat, err := strconv.ParseUint(parsedSpeed, 10, 32)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error, %v", err)
		}

		decodedData.Speed = uint16(parsedSpeedFloat)

		decoded.Data = append(decoded.Data, decodedData)

	}

	return decoded, nil
}
