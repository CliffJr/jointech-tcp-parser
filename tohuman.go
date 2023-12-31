package jointechparser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// toHumanReadable updates some fields in Decoded and returns human-readable data as JSON
func (d *Decoded) toHumanReadable() (Decoded, error) {
	// Update or modify fields as needed
	/*	d.ProtocolVersion = protocolVersion(d.ProtocolVersion)
		d.DeviceType = deviceType(d.DeviceType)
		d.DataType = dataType(d.DataType)
		d.Date = parseDate(d.Date)
		d.Time = parseTime(d.Time)

		directionIndicator, err := hexToByte(d.DirectionIndicator)
		if err != nil {
			return Decoded{}, fmt.Errorf("Error converting to Binary: %v", err)
		}

		fixedValue, longitude, latitude, positioning := decodeDirectionIndicator(directionIndicator)

		d.DirectionIndicator = fixedValue + "," + longitude + "," + latitude + "," + positioning

		mileage, err := hexToDecimal(d.Mileage)
		if err != nil {
			return Decoded{}, fmt.Errorf("Error converting to Decimal: %v", err)
		}

		d.Mileage = strconv.FormatInt(mileage, 10)

		deviceStatusParser := d.DeviceStatusParser

		d.DeviceStatus = deviceStatus(deviceStatusParser)

		d.GSMSignalQuality = GSMSignalQuality(d.GSMSignalQuality)

		//instance of ACLData
		decodedData := ACLData{}

		//standardize time
		decodedData.UtimeMs = toSeconds(d.Time)

		//standardize time ms
		decodedData.Utime = toMilliseconds(d.Time)

		//standardize no of the satellites
		decodedData.VisSat = d.Data[0].VisSat

		//standardize lat
		lat, err := parseLatLng(int(d.Data[0].Lat))
		if err != nil {
			return Decoded{}, fmt.Errorf("Error converting to Binary: %v", err)
		}

		decodedData.Lat = lat

		//standardize lng
		lng, err := parseLatLng(int(d.Data[0].Lng))
		if err != nil {
			return Decoded{}, fmt.Errorf("Error converting to Binary: %v", err)
		}

		decodedData.Lng = lng

		//standardize speed
		speed, err := parseSpeed(strconv.Itoa(int(d.Data[0].Speed)))
		decodedData.Speed = speed

		//standardize angle/direction
		angle, err := direction(strconv.Itoa(int(d.Data[0].Angle)))

		decodedData.Angle = int32(angle)

		d.Data[0] = decodedData*/

	return *d, nil
}

func protocolVersion(value string) string {
	if value == "19" {
		return "JT701D"
	}
	return "JT701"
}

func deviceType(value string) string {
	//Extract the first 4 bits (0.5 byte)
	intValue, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	firstHalf := intValue & 0x0F

	fmt.Printf("First 0.5 byte: %X\n", firstHalf)

	switch firstHalf {
	case 1:
		return "Regular rechargeable JT701"
	default:
		return "Unknown data type"
	}
}

func dataType(value string) string {
	// Extract the first 4 bits (0.5 byte)
	intValue, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	// Extract the second 4 bits (another 0.5 byte)
	secondHalf := (intValue >> 4) & 0x0F
	fmt.Printf("Second 0.5 byte: %X\n", secondHalf)
	switch secondHalf {
	case 1:
		return "Real-time position data"
	case 2:
		return "Alarm data"
	case 3:
		return "Blind area position data"
	case 4:
		return "Sub-new position data (newly added by JT701D)"
	default:
		return "Real-time position data"
	}
}

func parseDate(dateString string) string {
	layout := "020106" // DDMMYY layout

	// Parse the date string
	parsedDate, err := time.Parse(layout, dateString)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	utcTime := parsedDate.String()
	return utcTime
}

func decodeDirectionIndicator(value byte) (string, string, string, string) {
	// Extract individual bits using bitwise operations
	bit0 := (value & 0x01) == 0x01
	bit1 := (value & 0x02) == 0x02
	bit2 := (value & 0x04) == 0x04
	bit3 := (value & 0x08) == 0x08

	// Interpret the bits
	positioning := "GPS not positioning"
	if bit0 {
		positioning = "GPS positioning"
	}

	latitude := "north latitude"
	if !bit1 {
		latitude = "south latitude"
	}

	longitude := "east longitude"
	if !bit2 {
		longitude = "west longitude"
	}

	fixedValue := "fixed value.1"
	if !bit3 {
		fixedValue = "fixed value.0"
	}

	return fixedValue, longitude, latitude, positioning
}

func parseLatLng(latOrLng int) (float64, error) {
	degrees := latOrLng / 1000000
	minutes := (latOrLng % 1000000) / 10000
	decimalMinutes := float64(latOrLng%10000) / 10000.0
	final := float64(degrees) + (float64(minutes)+decimalMinutes)/60.0
	scale := math.Pow(10, float64(6))
	formatted := math.Round(final*scale) / scale
	return formatted, nil
}

func parseSpeed(value string) (float64, error) {
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

func hexToDecimal(hexStr string) (int64, error) {
	if hexStr == "" {
		return 0, fmt.Errorf("hexToDecimal: empty string provided")
	}

	decimal, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("hexToDecimal: %v", err)
	}

	return decimal, nil
}

// Convert hex string to byte
func hexToByte(hexString string) (byte, error) {
	var b byte
	n, err := fmt.Sscanf(hexString, "%02x", &b)
	if err != nil || n != 1 {
		return 0, fmt.Errorf("error converting hex to byte: %v", err)
	}

	return b, nil
}

// Convert hex string to binary
func hexToBinary(hexStr string) (string, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")

	intVal, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return "", err
	}

	binaryStr := strconv.FormatInt(intVal, 2)

	return binaryStr, nil
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

		if index < 0 {
			index = 0
		}

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
	index := (int(byteNum[0]-'0')-1)*8 + int(bitNum[3]-'0')

	if index < 0 {
		return 0
	}

	return index
}

func GSMSignalQuality(value uint8) uint8 {
	if value == 0 {
		return 99
	}
	return value
}

func parseTime(value string) string {
	seconds, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		_ = fmt.Errorf("converting error, %v", err)
	}

	duration := time.Duration(seconds) * time.Second
	utcTime := time.Now().UTC()
	resultTime := utcTime.Add(duration)

	// Format the time as "hh:mm:ss"
	resultFormatted := resultTime.Format("15:04:05")

	return resultFormatted
}

func toMilliseconds(timeString string) uint64 {
	parsedTime, err := time.Parse("15:04:05", timeString)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return 0
	}

	milliseconds := parsedTime.UnixNano() / int64(time.Millisecond)

	return uint64(milliseconds)
}

func toSeconds(timeString string) uint64 {
	parsedTime, err := time.Parse("15:04:05", timeString)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return 0
	}

	seconds := parsedTime.Unix()

	return uint64(seconds)
}
