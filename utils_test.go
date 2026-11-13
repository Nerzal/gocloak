package gocloak_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nerzal/gocloak/v14"
)

func TestPString(t *testing.T) {
	p := "test value"
	v := gocloak.PString(&p)
	assert.Equal(t, p, v)
}

func TestPStringNil(t *testing.T) {
	v := gocloak.PString(nil)
	assert.Empty(t, v)
}

func TestPBool(t *testing.T) {
	p := true
	v := gocloak.PBool(&p)
	assert.True(t, v)

	p = false
	v = gocloak.PBool(&p)
	assert.False(t, v)

	v = gocloak.PBool(nil)
	assert.False(t, v)
}

func TestPInt(t *testing.T) {
	p := 42
	v := gocloak.PInt(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)

	v = gocloak.PInt(nil)
	assert.Equal(t, int(0), v)
	assert.IsType(t, int(0), v)
}

func TestPInt32(t *testing.T) {
	var p int32 = 42
	v := gocloak.PInt32(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)

	v = gocloak.PInt32(nil)
	assert.Equal(t, int32(0), v)
	assert.IsType(t, int32(0), v)
}

func TestPInt64(t *testing.T) {
	var p int64 = 42
	v := gocloak.PInt64(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)

	v = gocloak.PInt64(nil)
	assert.Equal(t, int64(0), v)
	assert.IsType(t, int64(0), v)
}

func TestPFloat32(t *testing.T) {
	var p float32 = 42.42
	v := gocloak.PFloat32(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)

	v = gocloak.PFloat32(nil)
	assert.Equal(t, float32(0), v)
	assert.IsType(t, float32(0), v)
}

func TestPFloat64(t *testing.T) {
	p := 42.42
	v := gocloak.PFloat64(&p)
	assert.Equal(t, p, v)
	assert.IsType(t, p, v)

	v = gocloak.PFloat64(nil)
	assert.Equal(t, float64(0), v)
	assert.IsType(t, float64(0), v)
}

func TestNilOrEmptyArray(t *testing.T) {
	a := gocloak.NilOrEmptyArray([]string{"c", "d"})
	b := gocloak.NilOrEmptyArray([]string{"", "b"})
	c := gocloak.NilOrEmptyArray([]string{})
	assert.False(t, a)
	assert.True(t, b)
	assert.True(t, c)
}

func TestUtil_NilOrEmpty(t *testing.T) {
	t.Parallel()
	assert.True(t, gocloak.NilOrEmpty(nil))
	assert.True(t, gocloak.NilOrEmpty(new("")))
	assert.False(t, gocloak.NilOrEmpty(new("x")))
}
