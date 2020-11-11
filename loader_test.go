package envconfig

import (
	"os"
	"strings"
	"testing"
)

type customValue struct {
	v string
}

func (cv *customValue) Set(v string) error {
	cv.v = strings.Repeat(v, 3)
	return nil
}

type testSpec struct {
	Str string       `env:"TEST_STRING"`
	Val *customValue `env:"TEST_CUST_VALUE"`
}

func TestLoader(t *testing.T) {
	os.Setenv("TEST_STRING", "foo")
	os.Setenv("TEST_CUST_VALUE", "lorem")

	spec := testSpec{
		Val: &customValue{},
	}

	if err := Process(&spec, true); err != nil {
		t.Errorf("processing failed with error: %v", err)
	}

	if spec.Str != "foo" {
		t.Errorf("expected Str to be foo, got %s", spec.Str)
	}

	if spec.Val.v != "loremloremlorem" {
		t.Errorf("expected Val to be loremloremlorem, got %s", spec.Val.v)
	}
}
