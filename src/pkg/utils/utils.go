package utils

import (
	"fmt"
	"runtime"
	"strings"
)

func GetTypeString[G any]() string {
	fullType := fmt.Sprintf("%T", *new(G))
	lastDotIndex := strings.LastIndex(fullType, ".")
	if lastDotIndex == -1 {
		return fullType
	}

	return fullType[lastDotIndex+1:]
}

func GetSymbolName(skip int) string {
	// Get the program counter (pc) and other details from the runtime
	pc, _, _, _ := runtime.Caller(skip) // 1 means the caller of this function
	// Get the function details using the program counter
	fn := runtime.FuncForPC(pc)
	// Return the name of the function
	return fn.Name()
}

func GetFunctionName(full string) string {
	lastDotIndex := strings.LastIndex(full, ".")
	if lastDotIndex == -1 {
		return ""
	}

	// Extract the function name
	return full[lastDotIndex+1:]
}

func ExtractTypeName(full string) string {
	// Find the pointer notation "(*" and the closing parenthesis ")"
	start := strings.Index(full, "(*")
	end := strings.Index(full, ").")

	if start == -1 || end == -1 {
		return ""
	}

	// Extract the type name between "(*" and ")."
	return full[start+len("(*") : end]
}
