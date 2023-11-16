package jointechparser

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func packetReception(rawData string) string {
	// Split the raw data into packets using 0x24 as a delimiter
	packets := strings.Split(rawData, "24")

	// Process each packet
	var processedData []string
	for _, packet := range packets {
		// Trim leading and trailing spaces
		packet = strings.TrimSpace(packet)

		// Remove spaces between hex values
		packet = strings.ReplaceAll(packet, " ", "")

		// Check if the packet is not empty
		if packet != "" {
			// Convert hex string to bytes
			bytes, err := hex.DecodeString(packet)
			if err != nil {
				fmt.Println("Error decoding hex:", err)
				continue
			}

			for _, b := range bytes {
				fmt.Printf("%02X ", b)
			}
			fmt.Println()

			processedData = append(processedData, fmt.Sprintf("%02X", bytes))
		}
	}

	result := strings.Join(processedData, "\n")

	return result
}
