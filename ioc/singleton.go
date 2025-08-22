package ioc

import (
	"reflect"
)

func RegisterSingleton[T any](instance T) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := singletonServices[tp]; !ok {
		singletonServices[tp] = make(map[interface{}]service)
	}

	singletonServices[tp][0] = newService(instance, singleton)
}

func RegisterKeyedSingleton[T any](instance T, key interface{}) {
	tp := reflect.TypeOf((*T)(nil)).Elem()

	mu.Lock()
	defer mu.Unlock()

	if _, ok := singletonServices[tp]; !ok {
		singletonServices[tp] = make(map[interface{}]service)
	}

	singletonServices[tp][key] = newService(instance, singleton)
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
