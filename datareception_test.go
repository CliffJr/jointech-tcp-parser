package jointechparser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPacketReception(t *testing.T) {
	rawData := "24 75 00 31 36 20 19 12 00 34 15 07 20 21 49 53 22 34 97 50 11 35 50 36 4F 00 68 00 00 00 00 05 00 00 00 00 10 E0 4F 04 44 0B 32 1F 00 07 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 02 28 37 35 30 30 33 31 33 36 32 30 2C 40 4A 54 29 24 75 00 31 36 20 19 11 00 34 15 07 20 21 50 25 22 34 88 02 11 35 50 23 1F 00 89 00 00 00 00 05 00 00 00 00 00 E0 4F 04 44 0B 32 1F 00 07 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 60"
	//rawData := "24 75 00 31 36 02 19 11 00 34 17 07 20 01 58 25 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 30 24 75 00 31 36 02 19 11 00 34 17 07 20 01 58 35 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 31 24 75 00 31 36 02 19 11 00 34 17 07 20 01 58 45 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 32"
	//rawData := "24 75 00 31 36 02 19 13 00 34 17 07 20 01 58 25 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 30 24 75 00 31 36 02 19 13 00 34 17 07 20 01 58 35 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 31 24 75 00 31 36 02 19 13 00 34 17 07 20 01 58 45 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 32 24 75 00 31 36 02 19 13 00 34 17 07 20 01 58 55 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 33 24 75 00 31 36 02 19 13 00 34 17 07 20 01 59 05 22 33 67 43 11 40 15 83 5F 00 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 34 24 75 00 31 36 02 19 13 00 34 17 07 20 01 59 16 22 33 67 43 11 40 15 83 5F 17 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 35 24 75 00 31 36 02 19 13 00 34 17 07 20 01 59 26 22 33 67 43 11 40 15 83 5F 17 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 36 24 75 00 31 36 02 19 13 00 34 17 07 20 01 59 36 22 33 67 43 11 40 15 83 5F 17 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 37 24 75 00 31 36 02 19 13 00 34 17 07 20 01 59 46 22 33 67 43 11 40 15 83 5F 17 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 38 24 75 00 31 36 02 19 13 00 34 17 07 20 01 59 56 22 33 67 43 11 40 15 83 5F 17 80 00 00 00 02 04 00 00 00 00 20 E0 49 04 44 0B 32 1C 00 02 0F 0F 0F 0F 0F 0F 0F 0F 0F 0F 00 00 01 CC 00 39"

	expected := []string{"2475003136201912003415072021495322349750113550364F006800000000050000000010E04F04440B321F00070F0F0F0F0F0F0F0F0F0F000001CC000228373530303331333632302C404A5429", "2475003136201911003415072021502522348802113550231F008900000000050000000000E04F04440B321F00070F0F0F0F0F0F0F0F0F0F000001CC0060"}

	result, err := PacketReception(rawData)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)

}