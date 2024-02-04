# go-ecs

[![PkgGoDev](https://pkg.go.dev/badge/github.com/jabolopes/go-ecs)](https://pkg.go.dev/github.com/jabolopes/go-ecs)

This library provides a generic Entity Component System for developing
games.

```go
// Construction.
e := ecs.New()

// Create entities.
entity := e.Add()

// Delete entities.
e.Delete(entity)

// Set component in entity.
ecs.Set(e, MyComponent{value})
ecs.Set2(e, MyComponent{value1}, OtherComponent{value2})

// Get component(s) given entity ID.
c, ok := ecs.Get[MyComponent](e, entityId)

c1, c2, ok := ecs.Get2[MyComponent, OtherComponent](e, entityId)

// Unset (or remove) component from entity.
ecs.Unset[MyComponent](e, entityId)

// Iterate all entities that have a component.
for it := ecs.Iterate[MyComponent](e); ; {
  c, ok := it.Next()
  if !ok {
    break
  }

  // Do something with 'c'.
}

// Iterate all entities that have both components.
for it := ecs.Join[MyComponent, OtherComponent](e); ; {
  c1, c2, ok := it.Next()
  if !ok {
    break
  }

  // Do something with 'c1' and 'c2'.
}

// Iterate all entities that have all 3 components.
for it := ecs.Join3[MyC1, MyC2, MyC3](e) ; ; { ... }

// Get singleton entity with given component.
entityID, c, ok := e.IterateAny[MyComponent](e)

// Get singleton entity with both components.
entityID, c1, c2, ok := e.JoinAny[MyComponent, OtherComponents](e)
```
