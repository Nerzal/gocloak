package gocloak

// PString returns a string value from a pointer
func PString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// PBool returns a boolean value from a pointer
func PBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

// PInt returns an integer value from a pointer
func PInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

// PInt32 returns an int32 value from a pointer
func PInt32(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

// PInt64 returns an int64 value from a pointer
func PInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

// PFloat32 returns a float32 value from a pointer
func PFloat32(value *float32) float32 {
	if value == nil {
		return 0
	}
	return *value
}

// PFloat64 returns a float64 value from a pointer
func PFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

// NilOrEmpty returns true if string pointer is nil or the string is empty.
func NilOrEmpty(value *string) bool {
	return value == nil || len(*value) == 0
}

// NilOrEmptyArray returns true if the slice is empty or its first element is an empty string.
func NilOrEmptyArray(value []string) bool {
	if len(value) == 0 {
		return true
	}
	return value[0] == ""
}

// PStringSlice returns the slice as-is, or an empty slice if nil.
func PStringSlice(value []string) []string {
	if value == nil {
		return []string{}
	}
	return value
}

// NilOrEmptySlice returns true if the slice is nil or empty.
func NilOrEmptySlice(value []string) bool {
	return len(value) == 0
}
