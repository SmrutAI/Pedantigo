package deserialize

import "reflect"

// BuildNode returns a plan/cache node for typ, breaking type cycles with a
// back-edge. It registers a freshly-allocated (empty) node in inProgress BEFORE
// calling populate, so a nested reference back to typ (self-referential types,
// e.g. Node{Children []Node}) resolves to the same in-progress *N pointer rather
// than recursing forever. A type that appears more than once without a cycle
// (a DAG) also resolves to the single shared node, which is correct.
//
// alloc allocates a new zero node; populate fills the node's fields (and may
// recurse via BuildNode with the same inProgress map). Callers pass a fresh
// inProgress map at the top of a build.
func BuildNode[N any](
	typ reflect.Type,
	inProgress map[reflect.Type]*N,
	alloc func() *N,
	populate func(*N),
) *N {
	if n, ok := inProgress[typ]; ok {
		return n
	}
	n := alloc()
	inProgress[typ] = n
	populate(n)
	return n
}
