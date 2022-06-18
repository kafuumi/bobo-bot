package request

import "mime"

const (
	ApplicationJson       = "application/json"
	ApplicationUrlencoded = "application/x-www-form-urlencoded"
)

type ContentType struct {
	typ   string
	param map[string]string
}

func NewContentType(typ string, param map[string]string) *ContentType {
	return &ContentType{
		typ:   typ,
		param: param,
	}
}

func ParseContentType(s string) (*ContentType, error) {
	typ, param, err := mime.ParseMediaType(s)
	if err != nil {
		return nil, err
	}
	return &ContentType{
		typ:   typ,
		param: param,
	}, nil
}

func (c *ContentType) Format() string {
	return mime.FormatMediaType(c.typ, c.param)
}

func (c *ContentType) Type() string {
	return c.typ
}

func (c *ContentType) Param(name string) string {
	return c.param[name]
}
