package set

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(item T)         { s[item] = struct{}{} }
func (s Set[T]) Remove(item T)      { delete(s, item) }
func (s Set[T]) Exists(item T) bool { _, exists := s[item]; return exists }
