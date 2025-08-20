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
	Transient serviceType = iota
	Singleton
	Scoped
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
