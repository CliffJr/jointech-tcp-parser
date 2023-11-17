// Copyright 2023 Filip Kroča.
package jointechparser

import (
	"fmt"
	"github.com/CliffJr/b2n"
	"strconv"
)

// Decoded struct represent decoded E-Lock JointTech data structure with all ACL(Automated Container Lock) data as return from function Decode
type Decoded struct {
	ProtocolHeader        string
	ProtocolVersion       string
	IMEI                  string
	TerminalID            string
	Date                  string
	Time                  string
	DeviceType            string
	DataType              string
	DataLength            string
	DirectionIndicator    string
	Mileage               string
	BindVehicleID         string
	DeviceStatusParser    string
	DeviceStatus          DeviceStatuses
	BatteryLevel          uint8
	CellIdPositionCode    string
	GSMSignalQuality      uint8
	FenceAlarmID          uint8
	MNCHighByte           uint8
	ExpandedDeviceStatus  uint8
	ExpandedDeviceStatus2 uint8
	DataSerialNo          uint8
	Data                  ACLData // Slice with ACL data
}

// ACLData represent one block of data
type ACLData struct {
	UtimeMs  uint64    // Utime is Time in mili seconds
	Utime    uint64    // Utime is Time in seconds
	Priority uint8     // JT does not provide this value
	Lat      float64   // Latitude (between 850000000 and -850000000), fit float64
	Lng      float64   // Longitude (between 1800000000 and -1800000000), fit float64
	Altitude int16     // JT does not provide this value
	Angle    int32     // Direction in degrees from the JT docs In degrees
	VisSat   uint8     // The number of GPS satellites
	Speed    float64   // Speed in km/h
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
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type DeviceStatuses struct {
	baseStationPositioning     bool
	enterFenceAlarm            bool
	exitFenceAlarm             bool
	lockRopeCutAlarm           bool
	vibrationAlarm             bool
	platformACKCommandRequired bool
	lockRopeState              bool
	motorState                 bool
	longTimeUnlockingAlarm     bool
	wrongPasswordAlarm         bool
	swipeIllegalRFIDCardAlarm  bool
	lowBatteryAlarm            bool
	backCoverOpenedAlarm       bool
	backCoverStatus            bool
	motorStuckAlarm            bool
	reserved                   bool
}

// Decode takes a pointer to a slice of bytes with raw data and return Decoded struct
func Decode(bs *[]byte) (Decoded, error) {
	decoded := Decoded{}
	var err error

	// check for minimum packet size
	if len(*bs) < 45 {
		return Decoded{}, fmt.Errorf("Minimum packet size is 45 Bytes, got %v", len(*bs))
	}

	// check for JT packet validity
	if (*bs)[0] != 0x24 {
		return Decoded{}, fmt.Errorf("Probably not JT packet, trashed")
	}

	// decode and validate IMEI
	decoded.IMEI, err = b2n.ParseBs2String(bs, 48, 8)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine protocol header in packet
	decodedProtocolHeader, err := b2n.ParseBs2Uint8(bs, 0)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	decoded.ProtocolHeader = strconv.Itoa(int(decodedProtocolHeader))

	decoded.TerminalID, err = b2n.ParseBs2String(bs, 1, 5)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine protocol version in packet
	parsedProtocol, err := b2n.ParseBs2Uint8(bs, 6)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert uint64 error, %v", err)
	}

	decoded.ProtocolVersion = strconv.Itoa(int(parsedProtocol))

	// determine device type in packet
	decodedDeviceType, err := b2n.ParseBs2Uint8(bs, 7)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	decoded.DeviceType = strconv.Itoa(int(decodedDeviceType))

	// determine data type in packet
	decodedDataType, err := b2n.ParseBs2Uint8(bs, 7)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	decoded.DataType = strconv.Itoa(int(decodedDataType))

	// determine data length in packet
	decoded.DataLength, err = b2n.ParseBs2String(bs, 8, 2)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine date in packet
	decoded.Date, err = b2n.ParseBs2String(bs, 10, 3)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine time in packet
	decoded.Time, err = b2n.ParseBs2String(bs, 13, 3)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine direction indicator in packet
	decodedDirectionIndicator, err := b2n.ParseBs2String(bs, 20, 5)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	decoded.DirectionIndicator, err = cleanDirectionIndicator(decodedDirectionIndicator)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine mileage in packet
	parseMileage, err := b2n.ParseBs2String(bs, 27, 4)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	decoded.Mileage = parseMileage

	// determine bind vehicle id in packet
	decoded.BindVehicleID, err = b2n.ParseBs2String(bs, 32, 4)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// just a parser to determine device status in packet
	decodedDeviceStatus, err := b2n.ParseBs2String(bs, 36, 2)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	decoded.DeviceStatusParser = decodedDeviceStatus

	//just initialize for a now
	deviceStatuses := DeviceStatuses{}
	decoded.DeviceStatus = deviceStatuses

	// determine battery level in packet
	decoded.BatteryLevel, err = b2n.ParseBs2Uint8(bs, 38)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// determine Cell Id Position Code in packet
	decoded.CellIdPositionCode, err = b2n.ParseBs2String(bs, 39, 4)
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

	// make an instance for decoded data
	decodedData := ACLData{}

	// time record in ms has 8 Bytes
	parsedTime, err := b2n.ParseBs2String(bs, 13, 3)
	if err != nil {
		return Decoded{}, fmt.Errorf("Decode error, %v", err)
	}

	// Convert string to uint64
	parsedTimeUint64, err := strconv.ParseUint(parsedTime, 10, 64)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert uint64 error, %v", err)
	}

	decodedData.UtimeMs = parsedTimeUint64

	decodedData.Utime = uint64(decodedData.UtimeMs / 1000)

	// parse priority will be nil because JT does not provide that value
	decodedData.Priority = 0

	// parse lat and validate GPS
	parsedLat, err := b2n.ParseBs2String(bs, 16, 4)

	// Convert string to uint32
	parsedLatInt32, err := strconv.ParseUint(parsedLat, 10, 64)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert error, %v", err)
	}

	decodedData.Lat = float64(parsedLatInt32)

	if !(decodedData.Lat > -850000000 && decodedData.Lat < 850000000) {
		return Decoded{}, fmt.Errorf("Invalid Lat value, want lat > -850000000 AND lat < 850000000, got %v", decodedData.Lat)
	}

	// parse Lng and validate GPS
	parsedLng, err := b2n.ParseBs2String(bs, 20, 5)

	cleanedLng, err := cleanLng(parsedLng)

	// Convert string to uint32
	parsedLngInt32, err := strconv.ParseUint(cleanedLng, 10, 64)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert error, %v", err)
	}

	decodedData.Lng = float64(parsedLngInt32)

	if !(decodedData.Lng > -1800000000 && decodedData.Lng < 1800000000) {
		return Decoded{}, fmt.Errorf("Invalid Lat value, want lat > -1800000000 AND lat < 1800000000, got %v", decodedData.Lng)
	}

	// JT does not provide the Altitude
	decodedData.Altitude = 0

	// parse Angle
	parsedAngle, err := b2n.ParseBs2Int32TwoComplement(bs, 26)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert error, %v", err)
	}

	decodedData.Angle = parsedAngle

	if decodedData.Angle > 360 {
		return Decoded{}, fmt.Errorf("Invalid Angle value, want Angle <= 360, got %v", decodedData.Angle)
	}

	// parse num. of visible satellites VisSat
	parsedSatellites, err := b2n.ParseBs2Uint8(bs, 31)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert error, %v", err)
	}

	decodedData.VisSat = uint8(parsedSatellites)

	// parse Speed
	parsedSpeed, err := b2n.ParseBs2Uint8(bs, 6)
	if err != nil {
		return Decoded{}, fmt.Errorf("Convert error, %v", err)
	}

	decodedData.Speed = float64(parsedSpeed)

	decoded.Data = decodedData

	return decoded, nil
}

func cleanDirectionIndicator(hexString string) (string, error) {
	if len(hexString) == 0 {
		return "", fmt.Errorf("empty string")
	}

	lastLetter := string(hexString[len(hexString)-1])

	return lastLetter, nil
}

func cleanLng(hexString string) (string, error) {
	if len(hexString) == 0 {
		return "", fmt.Errorf("empty string")
	}

	numericPart := hexString[:len(hexString)-1]

	return numericPart, nil
}
