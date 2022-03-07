package utils

import (
	"net/url"
	"path"
)

// UrlJoin 连接url字符串，如果列表中存在绝对路径，以最后一个绝对路径为准，之前的值会抛弃。
// 结果中会舍弃尾部的'/'符号
func UrlJoin(urls ...string) (string, error) {
	var ret *url.URL = nil

	for _, v := range urls {
		u, err := url.Parse(v)
		// 忽略错误
		if err != nil {
			return "", err
		}
		if u.IsAbs() || ret == nil {
			ret = u
		} else {
			ret.Path = path.Join(ret.Path, u.Path)
			ret.Fragment = u.Fragment
			if ret.RawQuery == "" || u.RawQuery == "" {
				ret.RawQuery = ret.RawQuery + u.RawQuery
			} else {
				ret.RawQuery = ret.RawQuery + "&" + u.RawQuery
			}
		}
	}

	if ret == nil {
		return "", nil
	}
	return ret.String(), nil
}
