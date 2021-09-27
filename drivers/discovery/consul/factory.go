package consul

import (
	"reflect"

	"github.com/eolinker/eosc"
)

//Register 注册consul驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:discovery_consul", NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

//NewFactory 创建consul驱动工厂
func NewFactory() eosc.IProfessionDriverFactory {
	return &factory{}
}

//Create 创建consul驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		driver:     driverName,
		configType: reflect.TypeOf((*Config)(nil)),
		params:     params,
	}, nil
}
