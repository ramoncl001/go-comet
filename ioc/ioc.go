package ioc

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

var (
	errDependencyNotFound = errors.New("dependency not found")
)

type serviceType int

const (
	transient serviceType = iota
	singleton
	scoped
)

type service struct {
	value interface{}
	sType serviceType
}

func newService[T any](value T, t serviceType) service {
	return service{
		value: value,
		sType: t,
	}
}

var transientServices = make(map[reflect.Type]map[interface{}]service)
var singletonServices = make(map[reflect.Type]map[interface{}]service)
var scopedServices = make(map[reflect.Type]map[interface{}]service)

func Resolve[T any](ctx context.Context) (T, error) {
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

func ResolveKeyed[T any](ctx context.Context, key interface{}) (T, error) {
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

func resolve(ctx context.Context, t reflect.Type) (interface{}, error) {
	if _, ok := singletonServices[t][0]; ok {
		return resolveSingleton(t, 0)
	} else if _, ok := transientServices[t][0]; ok {
		return resolveTransient(ctx, t, 0)
	} else if _, ok := scopedServices[t][0]; ok {
		return resolveScoped(ctx, t, 0)
	}

	return nil, errDependencyNotFound
}

func resolveKeyed(ctx context.Context, t reflect.Type, key interface{}) (interface{}, error) {
	if _, ok := singletonServices[t][key]; ok {
		return resolveSingleton(t, key)
	} else if _, ok := transientServices[t][key]; ok {
		return resolveTransient(ctx, t, key)
	} else if _, ok := scopedServices[t][key]; ok {
		return resolveScoped(ctx, t, key)
	}

	return nil, errDependencyNotFound
}

var mu sync.RWMutex
