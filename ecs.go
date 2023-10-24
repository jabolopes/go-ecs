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

type remover interface {
	Remove(int)
}

// TODO: Replace pools map with array.
type ECS struct {
	defaultPageSize int
	nullKey         int
	pools           map[int]unsafe.Pointer
	removers        []remover
	idGenerator     int
}

func getPool[T any](e *ECS) (*sparseset.Set[T], bool) {
	set, ok := e.pools[getTypeId[T]()]
	if !ok {
		return nil, false
	}

	return (*sparseset.Set[T])(set), true
}

func initPool[T any](e *ECS) *sparseset.Set[T] {
	if p, ok := getPool[T](e); ok {
		return p
	}

	var t *T
	elemType := reflect.TypeOf(t)
	method, ok := elemType.MethodByName("Destroy")

	var pool *sparseset.Set[T]
	if ok {
		options := sparseset.Options[T]{
			func(value *T) {
				method.Func.Call([]reflect.Value{reflect.ValueOf(value)})
			},
		}

		pool = sparseset.NewWithOptions[T](e.defaultPageSize, e.nullKey, options)
	} else {
		pool = sparseset.New[T](e.defaultPageSize, e.nullKey)
	}

	e.pools[getTypeId[T]()] = unsafe.Pointer(pool)
	e.removers = append(e.removers, pool)
	return pool
}

func (e *ECS) Add() int {
	id := e.idGenerator
	e.idGenerator++
	return id
}

func (e *ECS) Remove(entityId int) {
	for _, remover := range e.removers {
		remover.Remove(entityId)
	}
}

func New() *ECS {
	return &ECS{
		4096,                     /* defaultPageSize */
		1 << 20,                  /* nullKey */
		map[int]unsafe.Pointer{}, /* pools */
		nil,                      /* removers */
		0,                        /* idGenerator */
	}
}

// TODO: Use 'Set' instead.
func Add[T any](e *ECS, entityId int) *T {
	return initPool[T](e).Add(entityId)
}

func Get[T any](e *ECS, entityId int) (*T, bool) {
	set, ok := getPool[T](e)
	if !ok {
		return nil, false
	}

	return set.Get(entityId)
}

func Get2[A, B any](e *ECS, entityId int) (*A, *B, bool) {
	set1, ok := getPool[A](e)
	if !ok {
		return nil, nil, false
	}

	set2, ok := getPool[B](e)
	if !ok {
		return nil, nil, false
	}

	return sparseset.Lookup(entityId, set1, set2)
}

func Get3[A, B, C any](e *ECS, entityId int) (*A, *B, *C, bool) {
	set1, ok := getPool[A](e)
	if !ok {
		return nil, nil, nil, false
	}

	set2, ok := getPool[B](e)
	if !ok {
		return nil, nil, nil, false
	}

	set3, ok := getPool[C](e)
	if !ok {
		return nil, nil, nil, false
	}

	return sparseset.Lookup3(entityId, set1, set2, set3)
}

func Set[A any](e *ECS, entityId int, a A) {
	*initPool[A](e).Add(entityId) = a
}

func Set2[A, B any](e *ECS, entityId int, a A, b B) {
	set1 := initPool[A](e)
	set2 := initPool[B](e)

	*set1.Add(entityId) = a
	*set2.Add(entityId) = b
}

func Set3[A, B, C any](e *ECS, entityId int, a A, b B, c C) {
	set1 := initPool[A](e)
	set2 := initPool[B](e)
	set3 := initPool[C](e)

	*set1.Add(entityId) = a
	*set2.Add(entityId) = b
	*set3.Add(entityId) = c
}

func Unset[T any](e *ECS, entityId int) {
	set, ok := getPool[T](e)
	if !ok {
		return
	}

	set.Remove(entityId)
}

func Iterate[A any](e *ECS) *sparseset.Iterator[A] {
	set, ok := getPool[A](e)
	if !ok {
		return sparseset.EmptyIterator[A]()
	}

	return sparseset.Iterate(set)
}

func Join[A, B any](e *ECS) *sparseset.JoinIterator[A, B] {
	set1, ok := getPool[A](e)
	if !ok {
		return sparseset.EmptyJoinIterator[A, B]()
	}

	set2, ok := getPool[B](e)
	if !ok {
		return sparseset.EmptyJoinIterator[A, B]()
	}

	return sparseset.Join(set1, set2)
}

func Join3[A, B, C any](e *ECS) *sparseset.Join3Iterator[A, B, C] {
	set1, ok := getPool[A](e)
	if !ok {
		return sparseset.EmptyJoin3Iterator[A, B, C]()
	}

	set2, ok := getPool[B](e)
	if !ok {
		return sparseset.EmptyJoin3Iterator[A, B, C]()
	}

	set3, ok := getPool[C](e)
	if !ok {
		return sparseset.EmptyJoin3Iterator[A, B, C]()
	}

	return sparseset.Join3(set1, set2, set3)
}

func Join4[A, B, C, D any](e *ECS) *sparseset.Join4Iterator[A, B, C, D] {
	set1, ok := getPool[A](e)
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	set2, ok := getPool[B](e)
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	set3, ok := getPool[C](e)
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	set4, ok := getPool[D](e)
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	return sparseset.Join4(set1, set2, set3, set4)
}

func GetOne[T any](e *ECS) (int, *T, bool) {
	iterator := Iterate[T](e)
	entityId, t, ok := iterator.Next()
	return entityId, t, ok
}

func JoinOne[A, B any](e *ECS) (int, *A, *B, bool) {
	iterator := Join[A, B](e)
	entityId, a, b, ok := iterator.Next()
	return entityId, a, b, ok
}

func SortStableFunc[T any](e *ECS, compare func(int, *T, int, *T) int) {
	set, ok := getPool[T](e)
	if !ok {
		return
	}

	sparseset.SortStableFunc(set, compare)
}
