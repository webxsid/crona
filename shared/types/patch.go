package types

type Patch[T any] struct {
	Set   bool
	Value *T
}
