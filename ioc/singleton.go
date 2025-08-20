package ioc

import (
	"context"
	"reflect"
)

func RegisterSingleton[T any](instance T) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := singletonServices[tp]; !ok {
		singletonServices[tp] = make(map[interface{}]service)
	}

	singletonServices[tp][0] = newService(instance, Singleton)
}

func RegisterKeyedSingleton[T any](instance T, key interface{}) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := singletonServices[tp]; !ok {
		singletonServices[tp] = make(map[interface{}]service)
	}

	singletonServices[tp][key] = newService(instance, Singleton)
}

func ResolveSingleton[T any](ctx context.Context) (T, error) {
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

func ResolveKeyedSingleton[T any](ctx context.Context, key interface{}) (T, error) {
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

func resolveSingleton(t reflect.Type, key interface{}) (interface{}, error) {
	mu.RLock()
	defer mu.RUnlock()

	instance, ok := singletonServices[t][key]
	if !ok {
		return nil, errDependencyNotFound
	}

	if instance.value != nil {
		return instance.value, nil
	}

	return nil, errDependencyNotFound
}
