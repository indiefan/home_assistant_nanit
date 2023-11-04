package message

import (
	"strconv"
	"time"
)

type UnixTime time.Time

// MarshalJSON is used to convert the timestamp to JSON
func (t UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *UnixTime) UnmarshalJSON(unixTimeBytes []byte) (err error) {
	unixTimeStr := string(unixTimeBytes)
	unixTimeInt, err := strconv.ParseInt(unixTimeStr, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(unixTimeInt, 0)
	return nil
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t UnixTime) Unix() int64 {
	return time.Time(t).Unix()
}

// Time returns the JSON time as a time.Time instance in UTC
func (t UnixTime) Time() time.Time {
	return time.Time(t).UTC()
}

// String returns t as a formatted string
func (t UnixTime) String() string {
	return t.Time().String()
}
