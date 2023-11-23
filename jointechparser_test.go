package jointechparser

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	hexData := "2480006200111911003418042116225922348310113550543F12980000002D060000000020E028109228661F00010000868822040248195F000001CC0156"

	expectedDecoded := Decoded{
		ProtocolHeader:     "36",
		ProtocolVersion:    "25",
		IMEI:               "868822040248195F",
		TerminalID:         "8000620011",
		Date:               "180421",
		Time:               "162259",
		DeviceType:         "17",
		DataType:           "17",
		DataLength:         "0034",
		DirectionIndicator: "F",
		Mileage:            "0000002D",
		BindVehicleID:      "00000000",
		DeviceStatusParser: "20E0",
		DeviceStatus: DeviceStatuses{
			baseStationPositioning:     false,
			enterFenceAlarm:            false,
			exitFenceAlarm:             false,
			lockRopeCutAlarm:           false,
			vibrationAlarm:             false,
			platformACKCommandRequired: false,
			lockRopeState:              false,
			motorState:                 false,
			longTimeUnlockingAlarm:     false,
			wrongPasswordAlarm:         false,
			swipeIllegalRFIDCardAlarm:  false,
			lowBatteryAlarm:            false,
			backCoverOpenedAlarm:       false,
			backCoverStatus:            false,
			motorStuckAlarm:            false,
			reserved:                   false,
		},
		BatteryLevel:          40,
		CellIdPositionCode:    "10922866",
		GSMSignalQuality:      0x1F,
		FenceAlarmID:          0x0,
		MNCHighByte:           0x00,
		ExpandedDeviceStatus:  0x01,
		ExpandedDeviceStatus2: 0x0,
		DataSerialNo:          0x0,

		Data: ACLData{

			UtimeMs:  162259,
			Utime:    0xa2,
			Priority: 0,
			Lat:      22348310,
			Lng:      113550543,
			Altitude: 0,
			Angle:    -1744830464,
			VisSat:   06,
			Speed:    25,
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

func TestDecodeHealthcheck(t *testing.T) {
	hexData := "28373030303331333333342C404A5429"

	byteData, err := hex.DecodeString(hexData)
	assert.NoError(t, err)
	assert.NotEmpty(t, byteData)

	// Pass the address of the byte slice to the Decode function
	decoded, err := Decode(&byteData)
	assert.Error(t, err)
	assert.Empty(t, decoded)

}
