package pinboard

import (
	"encoding/json"
	"fmt"
	"time"
)

type Post struct {
	URL         string    `json:"href"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Hash        string    `json:"hash"`
	Meta        string    `json:"meta"`
	Shared      bool      `json:"shared"`
	Tag         string    `json:"tag"`
	User        string    `json:"user"`
	Time        time.Time `json:"time"`
	Toread      bool      `json:"toread"`
	Others      int       `json:"others"`
}

func ParsePost(i interface{}) (*Post, error) {
	var data []byte
	var err error
	switch v := i.(type) {
	case string:
		data = []byte(v)
	case map[string]interface{}:
		data, err = json.Marshal(i)
		if err != nil {
			data = nil
		}
	}

	var p *Post

	if data != nil {
		if err = json.Unmarshal(data, &p); err != nil {
			p = nil
		}
	}

	if p == nil {
		return nil, fmt.Errorf("Cannot convert %T to Post", i)
	}

	return p, nil

}
