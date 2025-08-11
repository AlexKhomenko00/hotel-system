package shared

import (
	"strings"
	"time"
)

type Date time.Time

func (d *Date) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	t, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}
