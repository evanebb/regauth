package configuration

import "strings"

type errorCollection struct {
	errs []error
}

func newErrorCollection() *errorCollection {
	return &errorCollection{make([]error, 0)}
}

func (c *errorCollection) HasErrors() bool {
	return len(c.errs) > 0
}

func (c *errorCollection) Add(err error) {
	if err == nil {
		return
	}

	c.errs = append(c.errs, err)
}

func (c *errorCollection) Error() string {
	var s []string
	for _, err := range c.errs {
		s = append(s, err.Error())
	}

	return strings.Join(s, ", ")
}
