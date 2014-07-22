package pinboard

import (
	"fmt"
	"net/url"
)

type arguments map[string]interface{}

func (a *arguments) Encode() string {
	args := url.Values{}
	for k, v := range *a {
		args.Set(k, fmt.Sprintf("%v", v))
	}
	return args.Encode()
}
func (a *arguments) Set(key string, val interface{}) *arguments {
	(*a)[key] = val
	return a
}
