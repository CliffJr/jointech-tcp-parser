package jointechparser

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func PacketReception(rawData string) ([]string, error) {
	// Split the raw data into packets using 0x24 as a delimiter
	packets := strings.Split(rawData, "24")

	// Process each packet
	var result []string
	for _, packet := range packets {
		packet = strings.TrimSpace(packet)

		// Remove spaces between hex values
		packet = strings.ReplaceAll(packet, " ", "")

		// Check if the packet is not empty
		if packet != "" {
			_, err := hex.DecodeString(packet)
			if err != nil {
				return nil, fmt.Errorf("error decoding hex: %v", err)
			}

			result = append(result, packet)
		}
	}

	return result, nil
}
