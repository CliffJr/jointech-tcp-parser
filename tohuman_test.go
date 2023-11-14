package jointechparser

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToHumanRead(t *testing.T) {
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
	expected := "Real-time position data"

	result := dataType(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestParseDate(t *testing.T) {
	hexValue := "020106"
	expected := "Real-time position data"

	result := parseDate(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)

}

func TestDecodeDirectionIndicator(t *testing.T) {
	hexValue := "F"
	expected := "1111, east longitude, north latitude, GPS positioning"

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
	hexDirection := "12"
	expected := 304

	result, err := direction(hexDirection)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestHexToDecimal(t *testing.T) {
	hexValue := "0x24"
	expected := 304

	result, err := hexToDecimal(hexValue)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestHexToByte(t *testing.T) {
	hexValue := "0x24"
	expected := 304

	result, err := hexToByte(hexValue)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestHexToBinary(t *testing.T) {
	hexValue := "F"
	expected := 1111

	result, err := hexToBinary(hexValue)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestDeviceStatus(t *testing.T) {

	hexValue := "20E0"
	expected := 1111

	result := deviceStatus(hexValue)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}
