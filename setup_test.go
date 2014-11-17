package frat

import (
	"testing"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }


type TestSuite struct{}

var _ = Suite(&TestSuite{})

var key = []byte("asdf1234asdf1234")