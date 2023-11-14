package jointechparser

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecode(t *testing.T) {
	hexData := "2480006200111911003418042116225922348310113550543F12980000002D060000000020E028109228661F00010000868822040248195F000001CC0156"

	expectedDecoded := Decoded{
		ProtocolHeader:        36,
		ProtocolVersion:       "25",
		IMEI:                  "868822040248195F",
		TerminalID:            "8000620011",
		Date:                  "180421",
		DeviceType:            "1",
		DataType:              "1",
		DataLength:            "0034",
		DirectionIndicator:    "F",
		Mileage:               "0000002D",
		BindVehicleID:         "00000000",
		DeviceStatus:          "20E0",
		BatteryLevel:          40,
		CellIdPositionCode:    "10922866",
		GSMSignalQuality:      1,
		FenceAlarmID:          05,
		MNCHighByte:           00,
		ExpandedDeviceStatus:  01,
		ExpandedDeviceStatus2: 01,
		DataSerialNo:          86,

		Data: []ACLData{
			{
				UtimeMs:  162259,
				Utime:    162259,
				Priority: 0,
				Lat:      22348310,
				Lng:      113550543,
				Altitude: 0,
				Angle:    98,
				VisSat:   06,
				Speed:    12,
			},
		},
	}

	byteData, err := hex.DecodeString(hexData)
	assert.NoError(t, err)
	assert.NotEmpty(t, byteData)

	// Pass the address of the byte slice to the Decode function
	decoded, err := Decode(&byteData)
	assert.NoError(t, err)
	assert.NotEmpty(t, decoded)
	assert.Equal(t, expectedDecoded, decoded)

}
