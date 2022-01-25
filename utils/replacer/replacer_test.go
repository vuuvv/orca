package replacer

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestParse(t *testing.T) {
	template := `
select * from t_user as u
where ${userId:-1=1} and ${nickname:-1=1}
order by id desc
`
	r := New(template)
	r.Parse()

	if len(r.segments) != 3 {
		t.Errorf("segements lenght expect 3 but %d", len(r.segments))
	}

	first := `
select * from t_user as u
where `
	last := `
order by id desc
`
	if string(r.segments[0]) != first {
		t.Errorf("segements[0] expect '%s', but '%s'", first, r.segments[0])
	}
	if string(r.segments[1]) != " and " {
		t.Errorf("r.segments[1] should be ' and '")
	}
	if string(r.segments[2]) != last {
		t.Errorf("segements[2] expect '%s', but '%s'", last, r.segments[2])
	}

	if len(r.variables) != 2 {
		t.Errorf("variables length expect 2 but %d", len(r.variables))
	}

	var1 := r.variables[0]
	if var1.Name != "userId" {
		t.Errorf("var1 should be 'userId' but '%s'", var1.Name)
	}
	if var1.Default != "1=1" {
		t.Errorf("var1 should be '1=1' but '%s'", var1.Default)
	}

	var2 := r.variables[1]
	if var2.Name != "nickname" {
		t.Errorf("var1 should be 'nickname' but '%s'", var1.Name)
	}
	if var2.Default != "1=1" {
		t.Errorf("var1 should be '1=1' but '%s'", var1.Default)
	}
}

func TestReplace(t *testing.T) {
	template := `
select * from t_user as u
where ${userId:-1=1} and ${nickname:-1=1} and ${phone}
order by id desc
`
	r := New(template)

	s, _ := r.Replace(nil)

	assert.Equal(t, s, `
select * from t_user as u
where 1=1 and 1=1 and ${phone}
order by id desc
`, "replace with empty map")

	s, _ = r.Replace(map[string]string{
		"userId":   "?",
		"nickname": "?",
		"phone":    "1234567",
	})

	assert.Equal(t, s, `
select * from t_user as u
where ? and ? and 1234567
order by id desc
`, "replace with all variables")

}

func TestReplaceDuplicate(t *testing.T) {
	template := `
select * from t_user as u
where ${userId:-1=1} and ${userId:-1=1} and ${phone}
order by id desc
`
	r := New(template)

	s, _ := r.Replace(map[string]string{
		"userId": "vuuvv",
		"phone":  "1234567",
	})

	assert.Equal(t, s, `
select * from t_user as u
where vuuvv and vuuvv and 1234567
order by id desc
`, "replace with all variables")
}
