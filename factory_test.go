package proxy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFactory_Make(t *testing.T) {
	factory := NewFactory(map[string]*Config{
		"foo": {
			Host: "https://foo.com",
		},
		"bar": {
			Host: "https://bar.com",
		},
	})

	foo, err := factory.Make("foo")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, foo.config.Host, "https://foo.com")
	bar, err := factory.Make("bar")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, bar.config.Host, "https://bar.com")
	bar2, err := factory.Make("bar")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, bar, bar2)
}
