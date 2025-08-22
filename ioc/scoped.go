package ioc

import (
	"context"
	"reflect"
)

func RegisterScoped[T any](provider interface{}) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := scopedServices[tp]; !ok {
		scopedServices[tp] = make(map[interface{}]service)
	}
	scopedServices[tp][0] = newService(provider, scoped)
}

func RegisterKeyedScoped[T any](provider interface{}, key interface{}) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := scopedServices[tp]; !ok {
		scopedServices[tp] = make(map[interface{}]service)
	}
	scopedServices[tp][key] = newService(provider, scoped)
}

func resolveScoped(ctx context.Context, t reflect.Type, key interface{}) (interface{}, error) {
	mu.RLock()
	defer mu.RUnlock()

	service := ctx.Value(t)
	if service != nil {
		return service, nil
	}

	provider, ok := scopedServices[t][key]
	if !ok {
		return nil, errDependencyNotFound
	}

	tp := reflect.TypeOf(provider.value)
	if tp.Kind() != reflect.Func {
		return provider.value, nil
	}

	args := make([]reflect.Value, tp.NumIn())
	for i := 0; i < tp.NumIn(); i++ {
		argType := tp.In(i)
		arg, err := resolve(ctx, argType)
		if err != nil {
			return nil, err
		}
		args[i] = reflect.ValueOf(arg)
	}

	result := reflect.ValueOf(provider.value).Call(args)
	ctx = context.WithValue(ctx, t, result[0].Interface())

	return result[0].Interface(), nil
}
