package utils

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/stretchr/testify/assert"
	"github.com/vuuvv/vcommon/models"
	"math/rand"
	"testing"
)

func TestEncodeIntToBase64(t *testing.T) {
	//fmt.Println(1 << 8, 1 << 16, 1 << 24, 1 << 32, 1 << 40, 1 << 48, 1 << 56, int64(1) << 60)
	v1 := maxInt
	e1, err := EncodeIntToBase64Like(v1)
	assert.NoError(t, err, "解析出错")
	d1, err := DecodeBase64LikeToInt(e1)
	assert.NoError(t, err, "解析出错")
	assert.Equal(t, v1, d1)
}

func TestEncodeIntToBase64LikeRandom(t *testing.T) {
	for i := 0; i < 100; i++ {
		v := rand.Intn(1000)
		e, err := EncodeIntToBase64Like(v)
		assert.NoError(t, err, "解析出错")
		d, err := DecodeBase64LikeToInt(e)
		assert.NoError(t, err, "解析出错")
		assert.Equal(t, v, d)
		fmt.Println(v, e, d)
	}
	for i := 0; i < 100; i++ {
		v := rand.Intn(maxInt)
		e, err := EncodeIntToBase64Like(v)
		assert.NoError(t, err, "解析出错")
		d, err := DecodeBase64LikeToInt(e)
		assert.NoError(t, err, "解析出错")
		assert.Equal(t, v, d)
		fmt.Println(v, e, d)
	}
}

type EmbeddedA struct {
	models.Tree
	Name string
	Age  int
}

type EmbeddedB struct {
	ParentId int
	Name     string
	Age      int
}

func TestStructsEmbedded(t *testing.T) {
	a := &EmbeddedA{Name: "a", Age: 10}
	a.Tree.ParentId = 1
	b := &EmbeddedB{ParentId: 2, Name: "b", Age: 10}

	sa := structs.New(a)
	sb := structs.New(b)

	for _, v := range ExpandFields(sa.Fields()) {
		fmt.Println(v.Name())
	}
	for _, v := range sb.Fields() {
		fmt.Println(v.Name())
	}
}

func TestRand(t *testing.T) {
	fmt.Println(rand.Intn(2))
}
