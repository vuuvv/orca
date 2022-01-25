package utils

import "github.com/fatih/structs"

func ExpandFields(fields []*structs.Field) []*structs.Field {
	var ret []*structs.Field

	for _, v := range fields {
		if v.IsEmbedded() {
			ret = append(ret, ExpandFields(v.Fields())...)
		} else {
			ret = append(ret, v)
		}
	}

	return ret
}
