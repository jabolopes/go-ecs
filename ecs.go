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

func Init[T any](e *ECS) *sparseset.Set[T] {
	if p, ok := GetPool[T](e); ok {
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
	return pool
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

func Get[T any](e *ECS, entityId int) (*T, bool) {
	set, ok := e.pools[getTypeId[T]()]
	if !ok {
		return nil, false
	}

	return (*sparseset.Set[T])(set).Get(entityId)
}

func Get2[A, B any](e *ECS, entityId int) (*A, *B, bool) {
	set1, ok := e.pools[getTypeId[A]()]
	if !ok {
		return nil, nil, false
	}

	set2, ok := e.pools[getTypeId[B]()]
	if !ok {
		return nil, nil, false
	}

	return sparseset.Lookup(entityId, (*sparseset.Set[A])(set1), (*sparseset.Set[B])(set2))
}

func Get3[A, B, C any](e *ECS, entityId int) (*A, *B, *C, bool) {
	set1, ok := e.pools[getTypeId[A]()]
	if !ok {
		return nil, nil, nil, false
	}

	set2, ok := e.pools[getTypeId[B]()]
	if !ok {
		return nil, nil, nil, false
	}

	set3, ok := e.pools[getTypeId[C]()]
	if !ok {
		return nil, nil, nil, false
	}

	return sparseset.Lookup3(entityId, (*sparseset.Set[A])(set1), (*sparseset.Set[B])(set2), (*sparseset.Set[C])(set3))
}

func Remove[T any](e *ECS, entityId int) {
	typeId := getTypeId[T]()
	set, ok := e.pools[typeId]
	if !ok {
		return
	}

	(*sparseset.Set[T])(set).Remove(entityId)
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

func Join4[A, B, C, D any](e *ECS) *sparseset.Join4Iterator[A, B, C, D] {
	a, ok := e.pools[getTypeId[A]()]
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	b, ok := e.pools[getTypeId[B]()]
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	c, ok := e.pools[getTypeId[C]()]
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	d, ok := e.pools[getTypeId[D]()]
	if !ok {
		return sparseset.EmptyJoin4Iterator[A, B, C, D]()
	}

	return sparseset.Join4((*sparseset.Set[A])(a), (*sparseset.Set[B])(b), (*sparseset.Set[C])(c), (*sparseset.Set[D])(d))
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

func SortStableFunc[T any](e *ECS, compare func(int, *T, int, *T) bool) {
	set, ok := e.pools[getTypeId[T]()]
	if !ok {
		return
	}

	sparseset.SortStableFunc((*sparseset.Set[T])(set), compare)
}
