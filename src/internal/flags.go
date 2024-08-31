package internal

import "strings"

type ShareFlags []string

func (s *ShareFlags) String() string {
	return strings.Join(*s, ",")
}

func (s *ShareFlags) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}
