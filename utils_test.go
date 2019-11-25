package gocloak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringP(t *testing.T) {
	p := StringP("test value")
	assert.Equal(t, "test value", *p)
}
func TestPString(t *testing.T) {
	p := "test value"
	v := PString(&p)
	assert.Equal(t, p, v)
}

func TestPStringNil(t *testing.T) {
	v := PString(nil)
	assert.Equal(t, "", v)
}

func TestBoolP(t *testing.T) {
	p1 := BoolP(false)
	assert.False(t, *p1)
	p2 := BoolP(false)
	assert.False(t, *p1)
	assert.False(t, p1 == p2)

	p := BoolP(true)
	assert.True(t, *p)
}

func TestPBool(t *testing.T) {
	p := true
	v := PBool(&p)
	assert.True(t, v)

	p = false
	v = PBool(&p)
	assert.False(t, v)
}

func TestIntP(t *testing.T) {
	p := IntP(42)
	assert.Equal(t, 42, *p)
}

func TestInt32P(t *testing.T) {
	v := int32(42)
	p := Int32P(v)
	assert.Equal(t, v, *p)
}

func TestInt64P(t *testing.T) {
	v := int64(42)
	p := Int64P(v)
	assert.Equal(t, v, *p)
}

func TestPInt(t *testing.T) {
	var p int = 42
	v := PInt(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)
}

func TestPInt32(t *testing.T) {
	var p int32 = 42
	v := PInt32(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)
}

func TestPInt64(t *testing.T) {
	var p int64 = 42
	v := PInt64(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)
}

func TestFloat32P(t *testing.T) {
	v := float32(42.42)
	p := Float32P(v)
	assert.Equal(t, v, *p)
}
func TestFloat64P(t *testing.T) {
	v := float64(42.42)
	p := Float64P(v)
	assert.Equal(t, v, *p)
}

func TestPFloat32(t *testing.T) {
	var p float32 = 42.42
	v := PFloat32(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)
}
func TestPFloat64(t *testing.T) {
	var p float64 = 42.42
	v := PFloat64(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)
}
