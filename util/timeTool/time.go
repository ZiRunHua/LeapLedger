package timeTool

import (
	"encoding/json"
	"time"
)

type Timestamp struct{ time.Time }

func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{t}
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.Unix())
}
