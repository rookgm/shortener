package server

import "sync"

// valid content type
var validContentTypes = []string{"application/json", "text/html"}

type ContentTypeChecker struct {
	m    map[string]struct{}
	once sync.Once
}

// IsValid is checks the acceptable content type
func (ch *ContentTypeChecker) IsValid(s string) bool {
	ch.once.Do(func() {
		ch.m = make(map[string]struct{})
		for _, v := range validContentTypes {
			ch.m[v] = struct{}{}
		}
	})

	_, ok := ch.m[s]
	return ok
}
