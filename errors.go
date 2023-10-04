package sftpsender

import "strings"

type UpdateError struct {
	Errs []error
}

func (ue UpdateError) Error() string {
	var sb strings.Builder
	for _, err := range ue.Errs {
		sb.WriteString(err.Error())
		sb.WriteRune('\n')
	}
	result := sb.String()
	return result[:len(result)-1]
}
