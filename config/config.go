package config

import (
	"github.com/vuuvv/errors"
	"github.com/vuuvv/mapstructure"
	"github.com/vuuvv/viper"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Loader interface {
	Load(path string) error
	Get(name string) interface{}
	IsSet(key string) bool
	Unmarshal(output interface{}, names ...string) error
}

type ViperConfigLoader struct {
	viper *viper.Viper
}

type envKeyReplacer struct {
}

// Replace system environment is separate by char "_", and all uppercase
func (r *envKeyReplacer) Replace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(s), ".", "_"), "-", "_")
}

func NewViperConfigLoader() *ViperConfigLoader {
	val := &ViperConfigLoader{}
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(&envKeyReplacer{}),
	)
	v.AutomaticEnv()

	val.viper = v
	return val
}

func (v *ViperConfigLoader) Load(path string) (err error) {
	ext := filepath.Ext(path)[1:]
	v.viper.SetConfigType(ext)
	v.viper.SetConfigFile(path)
	return errors.WithStack(v.viper.ReadInConfig())
}

func (v *ViperConfigLoader) Get(name string) interface{} {
	return v.viper.Get(name)
}

func (v *ViperConfigLoader) IsSet(name string) bool {
	return v.viper.IsSet(name)
}

func (v *ViperConfigLoader) Unmarshal(output interface{}, names ...string) (err error) {
	name := strings.Join(names, ".")
	err = v.viper.UnmarshalKey(name, output, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			SystemEnvironmentHookFunc(name),
		)))
	return errors.WithStack(err)
}

func getEnv(name string) (val string, ok bool) {
	val, ok = os.LookupEnv(name)
	if ok {
		return
	}

	upperName := strings.ToUpper(name)
	val, ok = os.LookupEnv(name)
	if ok {
		return
	}

	val, ok = getEnvRelax(upperName)
	if ok {
		return
	}

	val, ok = getEnvRelax(name)
	if ok {
		return
	}
	return "", false
}

func MustLoad(path string) Loader {
	loader := NewViperConfigLoader()
	err := loader.Load(path)
	if err != nil {
		panic(err)
	}
	return loader
}

func getEnvRelax(name string) (val string, ok bool) {
	noDotName := strings.ReplaceAll(name, ".", "_")
	if name != noDotName {
		if val, ok = os.LookupEnv(noDotName); ok {
			return
		}
	}

	noHyphenName := strings.ReplaceAll(name, "-", "_")
	if name != noHyphenName {
		if val, ok = os.LookupEnv(noHyphenName); ok {
			return
		}
	}

	noDotNoHyphenName := strings.ReplaceAll(noHyphenName, ".", "_")
	if name != noHyphenName {
		if val, ok = os.LookupEnv(noDotNoHyphenName); ok {
			return
		}
	}
	return "", false
}

func SystemEnvironmentHookFunc(prefix ...string) mapstructure.DecodeHookFunc {
	p := ""
	if len(prefix) != 0 {
		p = strings.Join(prefix, "")
	}
	return func(name string, f reflect.Value, t reflect.Value) (interface{}, error) {
		// 环境变量不能设置结构体、map、slice、array
		if t.Kind() == reflect.Struct ||
			t.Kind() == reflect.Map ||
			t.Kind() == reflect.Slice ||
			t.Kind() == reflect.Array {
			return f.Interface(), nil
		}

		envName := name
		if len(name) != 0 {
			envName = p + "." + name
		}
		env, ok := getEnv(envName)
		if ok {
			return env, nil
		}
		return f.Interface(), nil
	}
}
