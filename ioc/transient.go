package ioc

import (
	"context"
	"reflect"
)

func RegisterTransient[T any](provider interface{}) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := transientServices[tp]; !ok {
		transientServices[tp] = make(map[interface{}]service)
	}
	transientServices[tp][0] = newService(provider, Transient)
}

func RegisterKeyedTransient[T any](provider interface{}, key interface{}) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := transientServices[tp]; !ok {
		transientServices[tp] = make(map[interface{}]service)
	}
	transientServices[tp][key] = newService(provider, Transient)
}

func ResolveTransient[T any](ctx context.Context) (T, error) {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	result, err := resolve(ctx, tp)
	if err != nil {
		return *new(T), err
	}

	if result == nil {
		return *new(T), errDependencyNotFound
	}

	return result.(T), nil
}

func ResolveKeyedTransient[T any](ctx context.Context, key interface{}) (T, error) {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	result, err := resolveKeyed(ctx, tp, key)
	if err != nil {
		return *new(T), err
	}

	if result == nil {
		return *new(T), errDependencyNotFound
	}

	return result.(T), nil
}

func resolveTransient(ctx context.Context, t reflect.Type, key interface{}) (interface{}, error) {
	mu.RLock()
	defer mu.RUnlock()

	provider, ok := transientServices[t][key]
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
	return result[0].Interface(), nil
}
