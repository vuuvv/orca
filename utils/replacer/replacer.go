package replacer

import (
	"bytes"
	"strings"
	"sync"
)

var cache sync.Map

type variable struct {
	Name       string
	Default    string
	HasDefault bool
}

func (v *variable) getDefault(prefix string, suffix string) string {
	if v.HasDefault {
		return v.Default
	} else {
		return prefix + v.Name + suffix
	}
}

func (v *variable) value(values map[string]string, prefix string, suffix string) (val string, exists bool) {
	if values == nil {
		return v.getDefault(prefix, suffix), false
	}
	if val, ok := values[v.Name]; ok {
		return val, true
	} else {
		return v.getDefault(prefix, suffix), false
	}
}

// Replacer 替换
// prefix,suffix必须为ascii字符串
type Replacer struct {
	template     string
	segments     [][]byte
	variables    []*variable
	prefix       string
	suffix       string
	escape       byte
	valueDelimit string
	once         sync.Once
}

var (
	defaultPrefix            = "${"
	defaultSuffix            = "}"
	defaultEscape       byte = '$'
	defaultValueDelimit      = ":-"
)

func New(template string) *Replacer {
	if val, ok := cache.Load(template); ok {
		return val.(*Replacer)
	}
	r := &Replacer{template: template}
	r.prefix = getDefaultString(r.prefix, defaultPrefix)
	r.suffix = getDefaultString(r.suffix, defaultSuffix)
	r.escape = getDefaultByte(r.escape, defaultEscape)
	r.valueDelimit = getDefaultString(r.valueDelimit, defaultValueDelimit)
	return r
}

func getDefaultString(value string, d string) string {
	if value == "" {
		return d
	}
	return value
}

func getDefaultByte(value byte, d byte) byte {
	if value == 0 {
		return d
	}
	return value
}

func match(buffer string, target string, pos int, bufferEnd int) int {
	length := len(target)
	// buffer 太短
	if pos+length > bufferEnd {
		return 0
	}
	if buffer[pos:pos+length] == target {
		return length
	}
	return 0
}

func (r *Replacer) addSegment(bytes []byte) {
	r.segments = append(r.segments, bytes)
}

func (r *Replacer) Parse() {
	template := r.template
	prefix := getDefaultString(r.prefix, defaultPrefix)
	suffix := getDefaultString(r.suffix, defaultSuffix)
	escape := getDefaultByte(r.escape, defaultEscape)

	pos := 0
	start := 0
	length := len(template)

	var buffer = &bytes.Buffer{}

	for {
		if pos >= length {
			r.addSegment(buffer.Bytes())
			break
		}
		prefixMatchLen := match(template, prefix, pos, length)
		if prefixMatchLen == 0 {
			buffer.WriteByte(template[pos])
			pos++
		} else {
			if pos > start && template[pos-1] == escape {
				// escape操作 如果前一个字节是escape字节则忽略前一个字符
				b := buffer.Bytes()
				buffer = bytes.NewBuffer(b[:len(b)-1])
			} else {
				// 匹配到开始标记

				// 开始标记和结束标记里的内容
				stmtBuffer := &bytes.Buffer{}

				// 开始寻找结束标记
				pos += prefixMatchLen

				for {
					if pos >= length {
						// 一直未找到结束标记，保留原文
						buffer.Write([]byte(prefix))
						buffer.Write(stmtBuffer.Bytes())
						r.addSegment(buffer.Bytes())
						break
					}
					suffixMatchLen := match(template, suffix, pos, length)
					if suffixMatchLen == 0 {
						stmtBuffer.WriteByte(template[pos])
						pos++
					} else {
						pos += suffixMatchLen
						// 将上一段字节数组保存
						r.addSegment(buffer.Bytes())
						buffer = &bytes.Buffer{}
						r.variables = append(r.variables, r.parseVariable(stmtBuffer.String()))
						break
					}
				}
			}
		}
	}
	cache.Store(r.template, r)
}

func (r *Replacer) parseVariable(stmt string) *variable {
	valueDelimit := r.valueDelimit
	ret := &variable{}
	i := strings.Index(stmt, valueDelimit)

	if i < 0 {
		ret.Name = stmt
		ret.HasDefault = false
	} else {
		ret.Name = stmt[:i]
		ret.Default = stmt[i+len(valueDelimit):]
		ret.HasDefault = true
	}
	return ret
}

func (r *Replacer) Replace(variables map[string]string) (str string, names []string) {
	r.once.Do(r.Parse)
	buf := bytes.Buffer{}
	rLen := len(r.segments)
	vLen := len(r.variables)
	length := rLen
	if vLen < rLen {
		length = vLen
	}
	for i := 0; i < length; i++ {
		buf.Write(r.segments[i])
		v := r.variables[i]
		s, exists := v.value(variables, r.prefix, r.suffix)
		buf.WriteString(s)
		if exists {
			names = append(names, v.Name)
		}
	}

	if vLen > rLen {
		v := r.variables[len(r.variables)-1]
		s, exists := v.value(variables, r.prefix, r.suffix)
		buf.WriteString(s)
		if exists {
			names = append(names, v.Name)
		}
	} else if rLen > vLen {
		buf.Write(r.segments[len(r.segments)-1])
	}
	return buf.String(), names
}
