package jointechparser

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToHumanReadable(t *testing.T) {
	decoded := Decoded{
		ProtocolHeader:        24,
		ProtocolVersion:       "JT701D",
		IMEI:                  "868822040248195F",
		TerminalID:            "8000620011",
		Date:                  "2006-01-02 00:00:00 +0000 UTC",
		DeviceType:            "Regular rechargeable JT701",
		DataType:              "Real-time position data",
		DataLength:            "52",
		DirectionIndicator:    "fixed value.1,east longitude,north latitude,GPS positioning",
		Mileage:               "45",
		BindVehicleID:         "00000000",
		DeviceStatus:          "20E0",
		BatteryLevel:          40,
		CellIdPositionCode:    "10922866",
		GSMSignalQuality:      31,
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

	// Call the toHumanReadable method
	humanReadable, err := decoded.toHumanReadable()
	assert.NoError(t, err)
	assert.NotEmpty(t, humanReadable)

	expectedDecoded := decoded

	// Convert the expected Decoded object to JSON
	expectedJSON, err := json.Marshal(expectedDecoded)
	assert.NoError(t, err)
	assert.NotEmpty(t, expectedDecoded)
	assert.Equal(t, expectedJSON, humanReadable)

}

func TestProtocolVersion(t *testing.T) {
	hexValue := "98"
	expected := "JT701D"

	result := protocolVersion(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestDeviceType(t *testing.T) {
	hexValue := "1"
	expected := "Regular rechargeable JT701"

	result := deviceType(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestDataType(t *testing.T) {
	hexValue := "1"
	expected := "Unknown data type"

	result := dataType(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestParseDate(t *testing.T) {
	hexValue := "020106"
	expected := "2006-01-02 00:00:00 +0000 UTC"

	result := parseDate(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)

}

func TestDecodeDirectionIndicator(t *testing.T) {
	hexValue := "F"
	expected := "fixed value.1,east longitude,north latitude,GPS positioning"

	directionIndicator, err := hexToByte(hexValue)
	assert.NoError(t, err)
	assert.NotEmpty(t, directionIndicator)

	fixedValue, longitude, latitude, positioning := decodeDirectionIndicator(directionIndicator)
	result := fixedValue + "," + longitude + "," + latitude + "," + positioning
	assert.NotEmpty(t, fixedValue)
	assert.NotEmpty(t, longitude)
	assert.NotEmpty(t, latitude)
	assert.NotEmpty(t, positioning)
	assert.Equal(t, expected, result)
}

func TestParseLatLng(t *testing.T) {
	lat := 22348310
	expected := 22.580517

	result, err := parseLatLng(lat)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestParseSpeed(t *testing.T) {
	hexSpeed := "12"
	expected := 33.3

	result, err := parseSpeed(hexSpeed)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestDirection(t *testing.T) {
	hexDirection := "98"
	expected := 304.0

	result, err := direction(hexDirection)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestHexToDecimal(t *testing.T) {
	hexValue := "0x24"
	expected := int64(36)

	result, err := hexToDecimal(hexValue)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestHexToByte(t *testing.T) {
	hexValue := "F"
	expected := uint8(0xf)

	result, err := hexToByte(hexValue)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestHexToBinary(t *testing.T) {
	hexValue := "F"
	expected := "1111"

	result, err := hexToBinary(hexValue)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestDeviceStatus(t *testing.T) {
	hexValue := "20E0"

	expected := "longTimeUnlockingAlarm: true " +
		"exitFenceAlarm: false " +
		"lockRopeState: true " +
		"motorState: true " +
		"swipeIllegalRFIDCardAlarm: false " +
		"backCoverStatus: false " +
		"baseStationPositioning: true " +
		"motorStuckAlarm: false " +
		"backCoverOpenedAlarm: false " +
		"reserved: false " +
		"platformACKCommandRequired: false " +
		"lowBatteryAlarm: false " +
		"wrongPasswordAlarm: false " +
		"lockRopeCutAlarm: false " +
		"enterFenceAlarm: false " +
		"vibrationAlarm: false "

	result := deviceStatus(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}
