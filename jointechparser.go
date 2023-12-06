// Copyright 2023 Filip Kroča.
package jointechparser

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/CliffJr/b2n"
)

const (
	// dec=15
	b2 = byte(0b00001111)
	// dec=240
	b1 = byte(0b11110000)
)

// Decoded struct represent decoded E-Lock JointTech root data structure with
// PAL (Positional / Alarm Lock) Data as return from Decode function
type Decoded struct {
	ProtocolHeader      string
	ProtocolVersion     string
	IMEI                string //15 digit IMEI in decimal format
	TerminalID          string //JointTech assigned ID in decimal format
	DeviceType          uint8
	DataType            uint8
	BindVehicleID       string
	ContainsHealthcheck bool
	Data                []PALData // Slice containing P(ositional)A(larm)L(ocation) data
}

type HighByteLockEvent byte

const (
	LongTimeUnlocking HighByteLockEvent = 1 << iota
	WrongPassword
	Swipe
	LowBattery
	CoverOpen
	CoverClosed
	MotorStuck
	Reserved
)

type LowByteLockEvent byte

const (
	BaseStationPositioning LowByteLockEvent = 1 << iota
	EnterFence
	ExitFence
	RopeCut
	Vibration
	AckRequired
	RopeInserted
	MotorLocked
)

// PALData represent a package of positional or alarm data recieved by a JT701D smart lock
type PALData struct {
	UtimeMs               uint64             // Utime is Time in mili seconds
	Utime                 uint64             // Utime is Time in seconds
	Lat                   float64            // Latitude (between 850000000 and -850000000), fits float64
	Lng                   float64            // Longitude (between 1800000000 and -1800000000), fits float64
	Angle                 int32              // Direction in degrees from the JT docs In degrees
	Distance              uint32             // Distance lock travelled in km (spec refers it as "Mileage")
	VisSat                uint8              // The number of GPS satellites
	Speed                 float64            // Speed in km/h
	DirectionIndicator    string             // Positioning - East/West longitude or North/South latitude
	Date                  string             // Package Date in DDMMYY format
	Time                  string             // Package Time in hh:mm:ss format
	BatteryLevel          uint8              // Percentage of available battery level
	CellIdPositionCode    uint32             // Most signifficant 2 bytes for Cell ID are stored first and lower 2 bytes for Pos code come next
	Mcc                   uint16             // Country code
	GSMSignalQuality      uint8              // A byte representing GSM signal strength (99 is for no signal detection)
	FenceAlarmID          uint8              // Entry and exit fence aleam ID (10 fences are supported)
	MNCHighByte           uint8              // Higher byte part of mobile operator code
	MNCLowByte            uint8              // Lower byte part of mobile operator code
	ExpandedDeviceStatus  uint8              // Contains wake up source with possible values from 0-10. Check page 15 of J701D Protocol Manual PDF for more details
	ExpandedDeviceStatus2 uint8              // For now holds info for battery charging status. Expected that JoinTech will place more status info in the future.
	SerialNo              uint8              // Sequence number of positional or alarm data recieved
	Length                uint16             // The data length from the date field to the data serial number in bytes
	HighEvents            *HighByteLockEvent // high bytes events (LongTimeUnlocking, WrongPassword, Swipe, LowBattery, CoverOpen, CoverClosed, MotorStuck, Reserved)
	LowEvents             *LowByteLockEvent  // low bytes events (BaseStationPositioning, EnterFence, ExitFence, RopeCut, Vibration, AckRequired, RopeInserted, MotorLocked)
}

// Returns mobile station ID
func (p *PALData) CellId() uint16 {
	// higher bits A = (N & 11110000) >> 4
	return uint16((p.CellIdPositionCode & 0xFFFF0000) >> 16)

}

// Return Location code "LAC"
func (p *PALData) LAC() uint16 {
	// lower bits B = N & 00001111
	return uint16(p.CellIdPositionCode & 0x0000FFFF)
}

func (k HighByteLockEvent) String() string {
	if k > LongTimeUnlocking {
		return fmt.Sprintf("<unknown key: %d>", k)
	}
	switch k {
	case Reserved:
		return "Reserved"
	case MotorStuck:
		return "MotorStuck"
	case CoverClosed:
		return "CoverClosed"
	case CoverOpen:
		return "CoverOpen"
	case LowBattery:
		return "LowBattery"
	case Swipe:
		return "Swipe"
	case WrongPassword:
		return "WrongPassword"
	case LongTimeUnlocking:
		return "LongTimeUnlocking"
	}

	// multiple keys
	var names []string
	for key := Reserved; key < LongTimeUnlocking; key <<= 1 {
		if k&key != 0 {
			names = append(names, key.String())
		}
	}
	return strings.Join(names, "|")
}

func (k LowByteLockEvent) String() string {
	if k > BaseStationPositioning {
		return fmt.Sprintf("<unknown key: %d>", k)
	}
	switch k {
	case MotorLocked:
		return "MotorLocked"
	case RopeInserted:
		return "RopeInserted"
	case AckRequired:
		return "AckRequired"
	case Vibration:
		return "Vibration"
	case RopeCut:
		return "RopeCut"
	case ExitFence:
		return "ExitFence"
	case EnterFence:
		return "EnterFence"
	case BaseStationPositioning:
		return "BaseStationPositioning"
	}

	// multiple keys
	var names []string
	for key := MotorLocked; key < BaseStationPositioning; key <<= 1 {
		if k&key != 0 {
			names = append(names, key.String())
		}
	}
	return strings.Join(names, "|")
}

func (p *PALData) AddHighEvent(key HighByteLockEvent) {
	(*p.HighEvents) |= key
}

func (p *PALData) HasHighEvent(key HighByteLockEvent) bool {
	return (*p.HighEvents)&key != 0
}

func (p *PALData) AddLowEvent(key LowByteLockEvent) {
	*p.LowEvents |= key
}

func (p *PALData) HasLowEvent(key LowByteLockEvent) bool {
	return *p.LowEvents&key != 0
}

// Decode takes a pointer to a slice of bytes with raw data and return Decoded struct
func Decode(bs *[]byte) (Decoded, error) {
	decoded := Decoded{}
	decoded.Data = make([]PALData, 0, 20)

	// check for minimum packet size - healthcheck has 16 bytes
	if len(*bs) < 16 {
		return Decoded{}, fmt.Errorf("Minimum packet size is 16 Bytes, got %v", len(*bs))
	}

	// check for JT packet validity
	if (*bs)[0] != 0x24 && (*bs)[0] != 0x28 {
		return Decoded{}, fmt.Errorf("%s Not a JT packet, trashed", hex.EncodeToString(*bs))
	}

	// 1- loop start wirh i=0, 2-loop start with i=62
outerLoop:
	for i := 0; i < len(*bs); {
		if (*bs)[i] == 0x28 {
			if len(*bs) > i+15 && (*bs)[i+13] == 0x4A && (*bs)[i+14] == 0x54 && (*bs)[i+15] == 0x29 {
				decoded.ContainsHealthcheck = true
				p := (*bs)[i+1 : i+11]
				decoded.TerminalID = *(*string)(unsafe.Pointer(&p))
				i = i + 16
				continue outerLoop
			}
			// TODO: implement parsing command output at this point
			if len(*bs) > i+15 && (*bs)[i+13] != 0x4A && (*bs)[i+14] != 0x54 && (*bs)[i+15] != 0x29 {
				for i < len(*bs) && (*bs)[i] != 0x29 {
					i++
				}
				i = i + 1
				continue outerLoop
			}

		}

		if i >= len(*bs) {
			break
		}

		//// determine protocol header in packet
		decodedProtocolHeader, err := b2n.ParseBs2Uint8(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error decodedProtocolHeader, %v", err)
		}
		decoded.ProtocolHeader = strconv.Itoa(int(decodedProtocolHeader))

		i = (i + 1) //1
		decoded.TerminalID, err = b2n.ParseBs2String(bs, i, 5)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error TerminalID, %v", err)
		}

		// determine protocol version in packet
		//i == 6
		i = (i + 5) //6
		parsedProtocol, err := b2n.ParseBs2Uint8(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert uint64 error parsedProtocol, %v", err)
		}

		decoded.ProtocolVersion = strconv.Itoa(int(parsedProtocol))

		// determine device type in packet
		i = (i + 1) //7
		decodedDeviceType, err := b2n.ParseBs2Uint8(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error decodedDeviceType, %v", err)
		}
		// higher bits A = (N & 11110000) >> 4
		decoded.DeviceType = (decodedDeviceType & b1) >> 4
		// lower bits B = N & 00001111
		decoded.DataType = decodedDeviceType & b2

		// make an struct for decoded PAL data
		decodedData := PALData{}
		i = (i + 1) //8
		decodedData.Length, err = b2n.ParseBs2Uint16(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error Length, %v", err)
		}

		// determine date in packet
		i = (i + 2) //10
		decodedData.Date, err = b2n.ParseBs2String(bs, i, 3)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error Date, %v", err)
		}
		i = (i + 3) //13
		// determine time in packet
		decodedData.Time, err = b2n.ParseBs2String(bs, i, 3)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error Time, %v", err)
		}
		i = (i + 3) //16

		// Convert string to uint64
		parsedTimeUint64, err := strconv.ParseUint(decodedData.Time, 10, 64)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert uint64 error parsedTimeUint64, %v", err)
		}

		decodedData.UtimeMs = parsedTimeUint64

		decodedData.Utime = uint64(decodedData.UtimeMs / 1000)
		//16
		parsedLat, err := b2n.ParseBs2String(bs, i, 4)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error parsedLat, %v", err)
		}
		//20
		i = i + 4
		// Convert string to uint32
		parsedLatInt32, err := strconv.ParseUint(parsedLat, 10, 64)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error parsedLatInt32, %v", err)
		}

		decodedData.Lat = float64(parsedLatInt32)

		if !(decodedData.Lat > -850000000 && decodedData.Lat < 850000000) {
			return Decoded{}, fmt.Errorf("Invalid Lat value, want lat > -850000000 AND lat < 850000000, got %v", decodedData.Lat)
		}

		// parse Lng and validate GPS
		//i=20
		parsedLng, err := b2n.ParseBs2String(bs, i, 5)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error parsedLng, %v", err)
		}

		cleanedLng, err := cleanLng(parsedLng)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error cleanedLng, %v", err)
		}
		//20
		i = i + 5

		// Convert string to uint32
		parsedLngInt32, err := strconv.ParseUint(cleanedLng, 10, 64)
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error, %v", err)
		}

		decodedData.Lng = float64(parsedLngInt32)

		if !(decodedData.Lng > -1800000000 && decodedData.Lng < 1800000000) {
			return Decoded{}, fmt.Errorf("Invalid Lat value, want lat > -1800000000 AND lat < 1800000000, got %v", decodedData.Lng)
		}

		decodedData.DirectionIndicator, err = cleanDirectionIndicator(parsedLng)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error, %v", err)
		}
		// parse Speed i=25
		parsedSpeed, err := b2n.ParseBs2Uint8(bs, i)
		//26
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error parsedSpeed, %v", err)
		}

		decodedData.Speed = float64(parsedSpeed) * 1.85
		// parse Angle (Direction)
		//i=26
		parsedAngle, err := b2n.ParseBs2Uint8(bs, i)
		//27
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error parsedAngle, %v", err)
		}

		decodedData.Angle = int32(parsedAngle) * 2

		if decodedData.Angle > 360 {
			return Decoded{}, fmt.Errorf("Invalid Angle value, want Angle <= 360, got %v", decodedData.Angle)
		}
		// determine mileage in packet
		//i = 27
		parsedMileage, err := b2n.ParseBs2Uint32(bs, i)
		//31
		i = i + 4
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error parsedMileage, %v", err)
		}

		decodedData.Distance = parsedMileage
		// parse num. of visible satellites VisSat
		//i=31
		parsedSatellites, err := b2n.ParseBs2Uint8(bs, i)
		//32
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Convert error parsedSatellites, %v", err)
		}

		decodedData.VisSat = uint8(parsedSatellites)
		// determine bind vehicle id in packet
		//i = 32
		decoded.BindVehicleID, err = b2n.ParseBs2String(bs, i, 4)
		//36
		i = i + 4
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error BindVehicleID, %v", err)
		}
		//i=36
		highByteStat, err := b2n.ParseBs2Uint8(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error highByteStat, %v", err)
		}
		hb := HighByteLockEvent(highByteStat)
		decodedData.HighEvents = &(hb)
		//37
		i = i + 1
		lowByteStat, err := b2n.ParseBs2Uint8(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error highByteStat, %v", err)
		}
		lb := LowByteLockEvent(lowByteStat)
		decodedData.LowEvents = &lb
		//38
		i = i + 1
		//i=38
		decodedData.BatteryLevel, err = b2n.ParseBs2Uint8(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error BatteryLevel, %v", err)
		}
		//39
		i = i + 1
		// determine Cell Id Position Code in packet
		//i=39
		decodedData.CellIdPositionCode, err = b2n.ParseBs2Uint32(bs, i)
		//43
		i = i + 4
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error, %v", err)
		}

		// determine GSM quality in packet
		//i=43
		decodedData.GSMSignalQuality, err = b2n.ParseBs2Uint8(bs, i)
		//44
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error, %v", err)
		}

		// determine Fence Alarm ID in packet
		//i=44
		decodedData.FenceAlarmID, err = b2n.ParseBs2Uint8(bs, i)
		//45
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error, %v", err)
		}

		// determine ExpandedDeviceStatus in packet
		//i=45
		decodedData.ExpandedDeviceStatus, err = b2n.ParseBs2Uint8(bs, i)
		//46
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error ExpandedDeviceStatus, %v", err)
		}

		// determine MNC High Byte in packet
		//i=46
		decodedData.MNCHighByte, err = b2n.ParseBs2Uint8(bs, i)
		//47
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error MNCHighByte, %v", err)
		}

		// determine ExpandedDeviceStatus2 in packet
		//i=47
		decodedData.ExpandedDeviceStatus2, err = b2n.ParseBs2Uint8(bs, i)
		//48
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error ExpandedDeviceStatus2, %v", err)
		}

		//48 868822040248195F 15 places it's not in hex it's in ascii format
		p := (*bs)[i : i+15]
		decoded.IMEI = *(*string)(unsafe.Pointer(&p))
		//63
		i = i + 15
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error IMEI, %v", err)
		}

		// Skip Cell ID in packet since it is part of CellIdPositionCode
		//i=63
		_, err = b2n.ParseBs2Uint16(bs, i)
		//65
		i = i + 2
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error CellId, %v", err)
		}

		// determine Mcc in packet
		//i=65
		decodedData.Mcc, err = b2n.ParseBs2Uint16(bs, i)
		//67
		i = i + 2
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error Mcc, %v", err)
		}

		// determine MNC Low Byte in packet
		//i=67
		decodedData.MNCLowByte, err = b2n.ParseBs2Uint8(bs, i)
		//68
		i = i + 1
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error MNCLowByte, %v", err)
		}

		// determine SerialNo of packet
		//i=68
		decodedData.SerialNo, err = b2n.ParseBs2Uint8(bs, i)
		if err != nil {
			return Decoded{}, fmt.Errorf("Decode error SerialNo, %v", err)
		}
		//69
		i = i + 1
		decoded.Data = append(decoded.Data, decodedData)
	}
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
