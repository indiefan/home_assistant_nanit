package utils

// ConstRefInt32 - constructs 32 bit integer
func ConstRefInt32(i int32) *int32 { return &i }

// ConstRefBool - constructs boolean
func ConstRefBool(b bool) *bool { return &b }

// ConstRefStr - construct string
func ConstRefStr(s string) *string { return &s }
