package ecs

import (
	"reflect"
	"unsafe"

	"github.com/jabolopes/go-sparseset"
)

var (
	typeIdGen = 0
	typeIds   = map[reflect.Type]int{}
)

func getTypeId[T any]() int {
	var t T
	typ := reflect.TypeOf(t)

	typeId, ok := typeIds[typ]
	if !ok {
		typeId = typeIdGen
		typeIds[typ] = typeId
		typeIdGen++
	}

	return typeId
}

// TODO: Replace pools map with array.
type ECS struct {
	defaultPageSize int
	nullKey         int
	pools           map[int]unsafe.Pointer
}

func New() *ECS {
	return &ECS{
		4096,                     /* defaultPageSize */
		1 << 20,                  /* nullKey */
		map[int]unsafe.Pointer{}, /* pools */
	}
}

func Init[T any](e *ECS, set *sparseset.Set[T]) {
	e.pools[getTypeId[T]()] = unsafe.Pointer(set)
}

func Add[T any](e *ECS, entityId int) *T {
	typeId := getTypeId[T]()
	set, ok := e.pools[typeId]
	if !ok {
		set := sparseset.New[T](e.defaultPageSize, e.nullKey)
		e.pools[typeId] = unsafe.Pointer(set)
		return set.Add(entityId)
	}

	return (*sparseset.Set[T])(set).Add(entityId)
}

func Remove[T any](e *ECS, entityId int) {
	typeId := getTypeId[T]()
	set, ok := e.pools[typeId]
	if !ok {
		return
	}

	(*sparseset.Set[T])(set).Remove(entityId)
}

func Get[T any](e *ECS, entityId int) (*T, bool) {
	set, ok := e.pools[getTypeId[T]()]
	if !ok {
		return nil, false
	}

	return (*sparseset.Set[T])(set).Get(entityId)
}

func GetPool[T any](e *ECS) (*sparseset.Set[T], bool) {
	set, ok := e.pools[getTypeId[T]()]
	if !ok {
		return nil, false
	}

	return (*sparseset.Set[T])(set), true
}

func Iterate[A any](e *ECS) *sparseset.Iterator[A] {
	set, ok := e.pools[getTypeId[A]()]
	if !ok {
		return sparseset.EmptyIterator[A]()
	}

	return sparseset.Iterate((*sparseset.Set[A])(set))
}

func Join[A, B any](e *ECS) *sparseset.JoinIterator[A, B] {
	a, ok := e.pools[getTypeId[A]()]
	if !ok {
		return sparseset.EmptyJoinIterator[A, B]()
	}

	b, ok := e.pools[getTypeId[B]()]
	if !ok {
		return sparseset.EmptyJoinIterator[A, B]()
	}

	return sparseset.Join((*sparseset.Set[A])(a), (*sparseset.Set[B])(b))
}

func Join3[A, B, C any](e *ECS) *sparseset.Join3Iterator[A, B, C] {
	a, ok := e.pools[getTypeId[A]()]
	if !ok {
		return sparseset.EmptyJoin3Iterator[A, B, C]()
	}

	b, ok := e.pools[getTypeId[B]()]
	if !ok {
		return sparseset.EmptyJoin3Iterator[A, B, C]()
	}

	c, ok := e.pools[getTypeId[C]()]
	if !ok {
		return sparseset.EmptyJoin3Iterator[A, B, C]()
	}

	return sparseset.Join3((*sparseset.Set[A])(a), (*sparseset.Set[B])(b), (*sparseset.Set[C])(c))
}
