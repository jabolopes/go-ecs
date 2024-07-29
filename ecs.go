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

func Init2[A, B any](e *ECS, entityId int, a A, b B) {
	e.Remove(entityId)
	Set2(e, entityId, a, b)
}

func Init3[A, B, C any](e *ECS, entityId int, a A, b B, c C) {
	e.Remove(entityId)
	Set3(e, entityId, a, b, c)
}

func Init4[A, B, C, D any](e *ECS, entityId int, a A, b B, c C, d D) {
	e.Remove(entityId)
	Set4(e, entityId, a, b, c, d)
}

func Init5[A, B, C, D, E any](ecs *ECS, entityId int, a A, b B, c C, d D, e E) {
	ecs.Remove(entityId)
	Set5(ecs, entityId, a, b, c, d, e)
}

func Init6[A, B, C, D, E, F any](ecs *ECS, entityId int, a A, b B, c C, d D, e E, f F) {
	ecs.Remove(entityId)
	Set6(ecs, entityId, a, b, c, d, e, f)
}

func Init7[A, B, C, D, E, F, G any](ecs *ECS, entityId int, a A, b B, c C, d D, e E, f F, g G) {
	ecs.Remove(entityId)
	Set7(ecs, entityId, a, b, c, d, e, f, g)
}

func Init8[A, B, C, D, E, F, G, H any](ecs *ECS, entityId int, a A, b B, c C, d D, e E, f F, g G, h H) {
	ecs.Remove(entityId)
	Set8(ecs, entityId, a, b, c, d, e, f, g, h)
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
func Set2[A, B any](ecs *ECS, entityId int, a A, b B) {
	*initPool[A](ecs).Add(entityId) = a
	*initPool[B](ecs).Add(entityId) = b
}

// Same as 'Set' for 3 component types.
func Set3[A, B, C any](ecs *ECS, entityId int, a A, b B, c C) {
	*initPool[A](ecs).Add(entityId) = a
	*initPool[B](ecs).Add(entityId) = b
	*initPool[C](ecs).Add(entityId) = c
}

// Same as 'Set' for 4 component types.
func Set4[A, B, C, D any](ecs *ECS, entityId int, a A, b B, c C, d D) {
	*initPool[A](ecs).Add(entityId) = a
	*initPool[B](ecs).Add(entityId) = b
	*initPool[C](ecs).Add(entityId) = c
	*initPool[D](ecs).Add(entityId) = d
}

// Same as 'Set' for 5 component types.
func Set5[A, B, C, D, E any](ecs *ECS, entityId int, a A, b B, c C, d D, e E) {
	*initPool[A](ecs).Add(entityId) = a
	*initPool[B](ecs).Add(entityId) = b
	*initPool[C](ecs).Add(entityId) = c
	*initPool[D](ecs).Add(entityId) = d
	*initPool[E](ecs).Add(entityId) = e
}

// Same as 'Set' for 6 component types.
func Set6[A, B, C, D, E, F any](ecs *ECS, entityId int, a A, b B, c C, d D, e E, f F) {
	*initPool[A](ecs).Add(entityId) = a
	*initPool[B](ecs).Add(entityId) = b
	*initPool[C](ecs).Add(entityId) = c
	*initPool[D](ecs).Add(entityId) = d
	*initPool[E](ecs).Add(entityId) = e
	*initPool[F](ecs).Add(entityId) = f
}

// Same as 'Set' for 7 component types.
func Set7[A, B, C, D, E, F, G any](ecs *ECS, entityId int, a A, b B, c C, d D, e E, f F, g G) {
	*initPool[A](ecs).Add(entityId) = a
	*initPool[B](ecs).Add(entityId) = b
	*initPool[C](ecs).Add(entityId) = c
	*initPool[D](ecs).Add(entityId) = d
	*initPool[E](ecs).Add(entityId) = e
	*initPool[F](ecs).Add(entityId) = f
	*initPool[G](ecs).Add(entityId) = g
}

// Same as 'Set' for 8 component types.
func Set8[A, B, C, D, E, F, G, H any](ecs *ECS, entityId int, a A, b B, c C, d D, e E, f F, g G, h H) {
	*initPool[A](ecs).Add(entityId) = a
	*initPool[B](ecs).Add(entityId) = b
	*initPool[C](ecs).Add(entityId) = c
	*initPool[D](ecs).Add(entityId) = d
	*initPool[E](ecs).Add(entityId) = e
	*initPool[F](ecs).Add(entityId) = f
	*initPool[G](ecs).Add(entityId) = g
	*initPool[H](ecs).Add(entityId) = h
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

func Join3Any[A, B, C any](e *ECS) (int, *A, *B, *C, bool) {
	iterator := Join3[A, B, C](e)
	entityId, a, b, c, ok := iterator.Next()
	return entityId, a, b, c, ok
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
