package utils

import (
	"fmt"
	"strings"

	_ "github.com/brendank310/aztui/pkg/logger"
)

func GetTypeString[G any]() string {
	fullType := fmt.Sprintf("%T", *new(G))
	lastDotIndex := strings.LastIndex(fullType, ".")
	if lastDotIndex == -1 {
		return fullType
	}

	return fullType[lastDotIndex+1:]
}
