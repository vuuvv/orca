package orm

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"reflect"
	"testing"
)

type testA struct {
	Name string
}

func (this testA) String() string {
	return this.Name
}

func testF[T any]() any {
	var t T
	return reflect.Zero(reflect.TypeOf(t)).Interface()
	//var i interface{}
	//i = t
	//return reflect.TypeOf(t)
}

func TestString(t *testing.T) {
	ss := Sequence{}
	t.Log(tableTitle(&ss))
	t.Log(fmt.Sprint(&ss))
	t.Log(testF[Sequence]())
	db, err := gorm.Open(postgres.Open("host=192.168.137.100 user=postgres password=1111aaaa dbname=unisoftcn_user port=5432 sslmode=disable TimeZone=Asia/Shanghai"), &gorm.Config{})
	if err != nil {
		t.Log(err)
		return
	}
	s, err := GetBy[Sequence](db, "key", "t_menu:DAAA")
	t.Log(s)
	if err != nil {
		t.Log(err)
		return
	}
	a := testA{Name: "a"}
	t.Log(a)
	t.Log(fmt.Sprintf("%s", a))
}
