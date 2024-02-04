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

// ECS is the Entity Component System.
//
// Several functions / methods return pointers to components. These pointers are
// only valid as long as the ECS is not modified by adding or removing
// entities. If adding or removing entities, the ECS's internal memory may
// reallocated, thus invaliding those pointers. For this reason, pointers to
// components obtained prior to adding or removing entities should not be
// accessed if the ECS is modified in this way. Also, it's recommended not to
// store pointers to components inside data types for later use because if they
// become invalidated it's easy to forget and access them later.
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

// Creates a new entity and returns the entity ID.
func (e *ECS) Add() int {
	id := e.idGenerator
	e.idGenerator++
	return id
}

// Removes an entity given its ID and removes all of its components.
func (e *ECS) Remove(entityId int) {
	for _, remover := range e.removers {
		remover.Remove(entityId)
	}
}

// Returns a new instance of the ECS with default options.
func New() *ECS {
	return &ECS{
		4096,                     /* defaultPageSize */
		1 << 20,                  /* nullKey */
		map[int]unsafe.Pointer{}, /* pools */
		nil,                      /* removers */
		0,                        /* idGenerator */
	}
}

// Initializes an entity and its component. If the entity already exists, it is
// first removed and then re-added. If the intention is not to initialize the
// entity, then use 'Set' instead.
func Init[A any](e *ECS, entityId int, a A) {
	e.Remove(entityId)
	*initPool[A](e).Add(entityId) = a
}

// Returns a component of the given type for an entity given its ID. Returns a
// pointer to the component and true if said entity exists, otherwise it returns
// false.
//
// The pointer is valid as long as the ECS is not modified (see ECS type)
func Get[T any](e *ECS, entityId int) (*T, bool) {
	set, ok := getPool[T](e)
	if !ok {
		return nil, false
	}

	return set.Get(entityId)
}

// Same as 'Get' for 2 component types. Returns true only if the entity has all
// types.
//
// The pointers are valid as long as the ECS is not modified (see ECS type)
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

// Same as 'Get' for 3 component types. Returns true only if the entity has all
// types.
//
// The pointers are valid as long as the ECS is not modified (see ECS type)
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

// Sets a component for an entity given its ID.
func Set[A any](e *ECS, entityId int, a A) {
	*initPool[A](e).Add(entityId) = a
}

// Same as 'Set' for 2 component types.
func Set2[A, B any](e *ECS, entityId int, a A, b B) {
	set1 := initPool[A](e)
	set2 := initPool[B](e)

	*set1.Add(entityId) = a
	*set2.Add(entityId) = b
}

// Same as 'Set' for 3 component types.
func Set3[A, B, C any](e *ECS, entityId int, a A, b B, c C) {
	set1 := initPool[A](e)
	set2 := initPool[B](e)
	set3 := initPool[C](e)

	*set1.Add(entityId) = a
	*set2.Add(entityId) = b
	*set3.Add(entityId) = c
}

// Removes a component from an entity given its ID. If the entity already does
// not have said component, then it's a no-op.
func Unset[T any](e *ECS, entityId int) {
	set, ok := getPool[T](e)
	if !ok {
		return
	}

	set.Remove(entityId)
}

// Returns an iterator that iterates all entities that have the given component
// type.
//
// for iterator := ecs.Iterate[MyComponent](e); ; {
//   c, ok := e.Next()
//   if !ok {
//     break
//   }
//
//   // Do something with 'c'.
// }
//
// The pointer returned by the iterator is valid as long as the ECS is not
// modified (see ECS type)
func Iterate[A any](e *ECS) *sparseset.Iterator[A] {
	set, ok := getPool[A](e)
	if !ok {
		return sparseset.EmptyIterator[A]()
	}

	return sparseset.Iterate(set)
}

// Returns an iterator that iterates all entities that have all component types.
//
// for iterator := ecs.Join[MyComponent, OtherComponent](e); ; {
//   c1, c2, ok := e.Next()
//   if !ok {
//     break
//   }
//
//   // Do something with 'c1' and 'c2'.
// }
//
// The pointers returned by the iterator are valid as long as the ECS is not
// modified (see ECS type)
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

// Same as 'Join' for 3 component types.
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

// Same as 'Join' for 4 component types.
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

// Returns any entity that has the given component. Returns the entity ID, the
// pointer to the component and true if said entity exists, otherwise it returns
// false.
//
// The pointer is valid as long as the ECS is not modified (see ECS type)
func IterateAny[T any](e *ECS) (int, *T, bool) {
	iterator := Iterate[T](e)
	entityId, t, ok := iterator.Next()
	return entityId, t, ok
}

// Returns any entity that has all the given components. Returns the entity ID,
// the pointers to the components and true if said entitiy exists, otherwise it
// returns false. The pointers are valid as long as the ECS is not modified.
//
// The pointers are valid as long as the ECS is not modified (see ECS type)
func JoinAny[A, B any](e *ECS) (int, *A, *B, bool) {
	iterator := Join[A, B](e)
	entityId, a, b, ok := iterator.Next()
	return entityId, a, b, ok
}

// Sorts the components using a stable sort function according to the given
// comparator function. The comparator function uses the same semantics are
// 'cmp.Compare' from the http://pkg.go.dev/cmp package.
func SortStableFunc[T any](e *ECS, compare func(int, *T, int, *T) int) {
	set, ok := getPool[T](e)
	if !ok {
		return
	}

	sparseset.SortStableFunc(set, compare)
}
