package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodec(t *testing.T) {
	type Book struct {
		Name string `json:"name"`
	}
	c := &codec{}
	v := []interface{}{1.0, "2", &Book{Name: "book"}}
	data, err := c.Marshal(v)
	assert.Equal(t, nil, err)

	rv := make([]interface{}, len(v))
	rv[2] = &Book{}
	err = c.Unmarshal(data, &rv)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1.0, rv[0].(float64))
	assert.Equal(t, "2", rv[1].(string))
	assert.Equal(t, "book", rv[2].(*Book).Name)
}
