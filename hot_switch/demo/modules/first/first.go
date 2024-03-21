package main

import "go-example/hot_switch/demo/modules/first/plugin"

func InvokeFunc(name string, params ...any) ([]any, error) {
	switch name {
	case "Add":
		a := params[0].(int)
		b := params[0].(int)
		return []any{plugin.Add(a, b)}, nil
	case "Sub":
		rs := []int{1, 2, 3}
		return []any{rs}, nil
	default:
		return nil, nil
	}
}
