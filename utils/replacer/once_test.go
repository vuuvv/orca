package replacer

import (
	"runtime"
	"sync"
	"testing"
)

func initialize() {
	value = &Replacer{}
}

var testOnce = &sync.Once{}

func GetWithOnce() *Replacer {
	testOnce.Do(initialize)
	return value
}

var (
	mu    = &sync.Mutex{}
	value *Replacer
)

func GetWithMutex() *Replacer {
	if value != nil {
		return value
	}
	mu.Lock()
	defer mu.Unlock()
	if value != nil {
		return value
	}
	initialize()
	return value
}

func BenchmarkOnce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetWithOnce()
	}
}

func BenchmarkMutex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetWithMutex()
	}
}

var gReplacer = New(`
select * from t_user as u
where ${userId:-1=1} and ${nickname:-1=1}
order by id desc
`)

var replacers = map[string]*Replacer{
	"github.com/vuuvv/vcommon/utils/replacer.BenchmarkTestTrace": gReplacer,
	gReplacer.template: gReplacer,
}

func trace2() string {
	pc := make([]uintptr, 15)
	runtime.Callers(2, pc)
	//n := runtime.Callers(2, pc)
	//frames := runtime.CallersFrames(pc[:n])
	//frame, _ := frames.Next()
	//return frame.Function
	return ""
	//fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
}

func TestTrace(t *testing.T) {
	key := trace2()
	if r, ok := replacers[key]; ok {
		r.Replace(map[string]string{
			"userId":   "?",
			"nickname": "?",
			"phone":    "1234567",
		})
	}
}

func BenchmarkTestTraceSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if r, ok := replacers[gReplacer.template]; ok {
			r.Replace(map[string]string{
				"userId":   "?",
				"nickname": "?",
				"phone":    "1234567",
			})
		}
	}
}

func BenchmarkTestTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trace2()
		if r, ok := replacers["github.com/vuuvv/vcommon/utils/replacer.BenchmarkTestTrace"]; ok {
			r.Replace(map[string]string{
				"userId":   "?",
				"nickname": "?",
				"phone":    "1234567",
			})
		}
	}
}

func parseReplacer() {
	template := `
select * from t_user as u
where ${userId:-1=1} and ${nickname:-1=1}
order by id desc
`
	r := New(template)
	r.Replace(map[string]string{
		"userId":   "?",
		"nickname": "?",
		"phone":    "1234567",
	})
}

func BenchmarkParseReplacer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseReplacer()
	}
}
