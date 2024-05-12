package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserValid(t *testing.T) {
	falseInfo := User{
		Username: "xxxx",
		Email:    "xxxxxx.net",
		Phone:    "xxxx",
		Avatar:   "xxxx",
		Password: "the_wonder_1243",
	}
	ok, err := falseInfo.Valid()
	assert.False(t, ok)
	assert.NotNil(t, err)

	normalInfo := User{
		Username: "abcde",
		Email:    "xxxxxx@net",
		Phone:    "xxxx",
		Avatar:   "xxxx",
		Password: "the_wonder_1243",
	}
	o, e := normalInfo.Valid()
	assert.True(t, o)
	assert.Nil(t, e)
}
