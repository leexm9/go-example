package hot_switch

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"plugin"
	"reflect"
)

var NotExistErr = errors.New("symbol does not exist")

type Plugin struct {
	Name      string
	File      string
	FileSha1  [sha1.Size]byte
	Remark    string
	unchanged bool

	p          *plugin.Plugin
	InvokeFunc func(name string, params ...any) ([]any, error)
	reloadable bool
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (pl *Plugin) Lookup(symName string, out any) error {
	if symName == "" {
		return fmt.Errorf("symName cannot be empty. dy: %s", pl.Name)
	}
	if isNil(out) {
		return fmt.Errorf("out cannot be nil. dy: %s, symName: %s", pl.Name, symName)
	}

	outVal := reflect.ValueOf(out)
	if k := outVal.Type().Kind(); k != reflect.Ptr {
		return fmt.Errorf("out must be a pointer. dy: %s, symName: %s", pl.Name, symName)
	}

	sym, err := pl.p.Lookup(symName)
	if err != nil {
		return NotExistErr
	}

	symVal := reflect.ValueOf(sym)
	symType := symVal.Type()
	elem := outVal.Elem()
	elemType := elem.Type()

	switch {
	case symType.AssignableTo(elemType):
		elem.Set(symVal)
		return nil
	case symType.Kind() == reflect.Ptr && symType.Elem().AssignableTo(elemType):
		elem.Set(symVal.Elem())
		return nil
	default:
		return fmt.Errorf("failed to assign %s to out. dy: %s, symType: %s, outType: %s",
			symName, pl.Name, symType, outVal.Type().String())
	}
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	switch vv := reflect.ValueOf(v); vv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Slice, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface:
		return vv.IsNil()
	default:
		return false
	}
}
