package jointechparser

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// toHumanReadable updates some fields in Decoded and returns human-readable data as JSON
func (d *Decoded) toHumanReadable() (string, error) {
	// Update or modify fields as needed
	//type Decoded struct {
	//	ProtocolHeader        uint8
	//	ProtocolVersion       uint8
	//	IMEI                  string
	//	TerminalID            string
	//	Date                  string
	//	DeviceType            string
	//	DataType              string
	//	DataLength            string
	//	DirectionIndicator    string
	//	Mileage               string
	//	BindVehicleID         string
	//	DeviceStatus          string
	//	BatteryLevel          uint8
	//	CellIdPositionCode    string
	//	GSMSignalQuality      uint8
	//	FenceAlarmID          uint8
	//	MNCHighByte           uint8
	//	ExpandedDeviceStatus  uint8
	//	ExpandedDeviceStatus2 uint8
	//	DataSerialNo          uint8
	//	Data                  []ACLData // Slice with ACL data
	//}
	d.DeviceType = "Updated Device Type"
	d.DataType = "Updated Data Type"

	// Convert Decoded to JSON
	jsonData, err := json.Marshal(d)
	if err != nil {
		return "", fmt.Errorf("Error marshalling to JSON: %v", err)
	}

	// Return the JSON string
	return string(jsonData), nil
}

func protocolVersion(value uint8) string {
	if value == 19 {
		return "JT701D"
	}
	return "JT701"
}

func latLng(latOrLng int) (float64, error) {
	degrees := latOrLng / 1000000
	minutes := (latOrLng % 1000000) / 10000
	decimalMinutes := float64(latOrLng%10000) / 10000.0
	final := float64(degrees) + (float64(minutes)+decimalMinutes)/60.0
	scale := math.Pow(10, float64(6))
	formatted := math.Round(final*scale) / scale
	return formatted, nil
}

func batteryLevel(value string) (float64, error) {
	decimalValue, err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	return float64(decimalValue), nil
}

func speed(value string) (float64, error) {
	decimalValue, err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	speed := float64(decimalValue) * 1.85
	scale := math.Pow(10, float64(6))
	formatted := math.Round(speed*scale) / scale

	return formatted, nil
}

func direction(value string) (float64, error) {
	decimalValue, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	direction := decimalValue * 2

	return float64(direction), nil
}

func deviceType(value string) int64 {
	// Extract the first 4 bits (0.5 byte)
	intValue, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	firstHalf := intValue & 0x0F

	fmt.Printf("First 0.5 byte: %X\n", firstHalf)
	return firstHalf
}

func dataType(value string) int64 {
	// Extract the first 4 bits (0.5 byte)
	intValue, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	// Extract the second 4 bits (another 0.5 byte)
	secondHalf := (intValue >> 4) & 0x0F
	fmt.Printf("Second 0.5 byte: %X\n", secondHalf)
	return secondHalf
}

func hexToBinary(hexStr string) (string, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")

	intVal, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return "", err
	}

	binaryStr := strconv.FormatInt(intVal, 2)

	return binaryStr, nil
}

func deviceStatus(value string) {
	binaryStr, err := hexToBinary(value)
	if err != nil {
		fmt.Println("Error:", err)
	}

	deviceStatus := parseDeviceStatus(binaryStr)

	for state, value := range deviceStatus {
		fmt.Printf("%s: %v\n", state, value)
	}

}

func parseDeviceStatus(binaryStr string) map[string]bool {
	deviceStatus := make(map[string]bool)

	// Reverse the binary string to match the description
	reversedBinary := reverseString(binaryStr)

	statesAndAlarms := map[string]string{
		"Byte1.BIT0": "baseStationPositioning",
		"Byte1.BIT1": "enterFenceAlarm",
		"Byte1.BIT2": "exitFenceAlarm",
		"Byte1.BIT3": "lockRopeCutAlarm",
		"Byte1.BIT4": "vibrationAlarm",
		"Byte1.BIT5": "platformACKCommandRequired",
		"Byte1.BIT6": "lockRopeState",
		"Byte1.BIT7": "motorState",
		"Byte2.BIT0": "longTimeUnlockingAlarm",
		"Byte2.BIT1": "wrongPasswordAlarm",
		"Byte2.BIT2": "swipeIllegalRFIDCardAlarm",
		"Byte2.BIT3": "lowBatteryAlarm",
		"Byte2.BIT4": "backCoverOpenedAlarm",
		"Byte2.BIT5": "backCoverStatus",
		"Byte2.BIT6": "motorStuckAlarm",
		"Byte2.BIT7": "reserved",
	}

	for bit, description := range statesAndAlarms {
		index := len(reversedBinary) - bitToInt(bit) - 1
		deviceStatus[description] = reversedBinary[index] == '1'
	}

	return deviceStatus
}

func reverseString(str string) string {
	reversed := []rune(str)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return string(reversed)
}

func bitToInt(bit string) int {
	bit = strings.TrimPrefix(bit, "Byte")
	parts := strings.Split(bit, ".")
	byteNum := parts[0]
	bitNum := parts[1]
	return (int(byteNum[0]-'0')-1)*8 + int(bitNum[3]-'0')
}
