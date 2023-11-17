package jointechparser

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// PacketReception TODO: Clean the packet by removing unnecessary clubbed information in future
func PacketReception(rawData string) ([]string, error) {
	// Split the raw data into packets using 0x24 as a delimiter
	packets := strings.Split(rawData, "24")

	// Process each packet
	var result []string
	for _, packet := range packets {
		packet = strings.TrimSpace(packet)
		packet = strings.ReplaceAll(packet, " ", "")

		if packet != "" {
			_, err := hex.DecodeString(packet)
			if err != nil {
				return nil, fmt.Errorf("error decoding hex: %v", err)
			}

			result = append(result, "24"+packet)
		}
	}

	return result, nil
}
