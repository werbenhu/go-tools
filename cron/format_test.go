package cron

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse1(t *testing.T) {
	pattern1 := NewFormat()
	pattern1.Second("01")
	pattern1.Second("02")
	ret1 := pattern1.Parse()
	expected1 := "1,2 * * * * ?"
	assert.Equal(t, expected1, ret1, "FAILED TestParse1 pattern1 expected:%s, actual::%s.", expected1, ret1)

	pattern2 := NewFormat()
	pattern2.Week("1")
	pattern2.Week("0")
	ret2 := pattern2.Parse()
	expected2 := "* * * * * 1,0"
	assert.Equal(t, expected2, ret2, "FAILED TestParse1 pattern2 expected:%s, actual::%s.", expected2, ret2)

	pattern3 := NewFormat()
	pattern3.Second("0")
	pattern3.Minute("1")
	pattern3.Hour("1")
	pattern3.Day("1")
	pattern3.Week("1")
	ret3 := pattern3.Parse()
	expected3 := "0 1 1 1 * 1"
	assert.Equal(t, expected3, ret3, "FAILED TestParse1 pattern2 expected:%s, actual::%s.", expected3, ret3)

	pattern4 := NewFormat()
	pattern4.Second("0")
	pattern4.Second("1")
	pattern4.Minute("1")
	pattern4.Minute("2")
	pattern4.Hour("1")
	pattern4.Hour("2")
	pattern4.Day("1")
	pattern4.Day("3")
	pattern4.Week("1")
	pattern4.Week("4")
	ret4 := pattern4.Parse()
	expected4 := "0,1 1,2 1,2 1,3 * 1,4"
	assert.Equal(t, expected4, ret4, "FAILED TestParse1 pattern2 expected:%s, actual::%s.", expected4, ret4)
}

func TestAddError(t *testing.T) {
	pattern := NewFormat()

	err := pattern.Hour("60")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddHour(60) test err:%s.", err)

	err = pattern.Minute("60")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddMin(60) test err:%s.", err)

	err = pattern.Second("60")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddSecond(60) test err:%s.", err)

	err = pattern.Hour("-1")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddHour(-1) test err:%s.", err)

	err = pattern.Minute("-1")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddMin(-1) test err:%s.", err)

	err = pattern.Second("-1")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddSecond(-1) test err:%s.", err)

	err = pattern.Hour("abc")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddHour(abc) test err:%s.", err)

	err = pattern.Minute("abc")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddMin(abc) test err:%s.", err)

	err = pattern.Second("abc")
	assert.NotEqual(t, nil, err, "FAILED TestAddError AddSecond(abc) test err:%s.", err)
}
