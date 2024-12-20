package validate

type Validatable interface {
	ValidationSpec() Spec
}

// func Validate[T any](obj T, specProviders ...func(T) Spec) error {
// 	var spec Spec
// 	if _, is := any(obj).(Validatable); is {
// 		spec = obj.(Validatable).ValidationSpec()
// 	}

// 	if len(specProviders) > 0 {
// 		spec = specProviders[0](obj)
// 	}

// 	if len(spec) == 0 {
// 		return nil // no validation spec so no action
// 	}

// 	var err error
// 	for _, s := range spec {
// 		for _, pred := range s.Predicates {
// 			thisErr := pred(s)
// 			if thisErr != nil {
// 				if err == nil {
// 					err = thisErr
// 				} else {
// 					err = fmt.Errorf("%s: %w", err.Error(), thisErr)
// 				}
// 			}
// 		}
// 	}

// 	return err
// }

type Spec = []fieldSpec

type fieldSpec struct {
	Field      any
	Predicates []Predicate
}

func FieldSpec(value any, predicates ...Predicate) fieldSpec {
	return fieldSpec{Field: value, Predicates: predicates}
}

type Predicate func(v any) error

// type User struct {
// 	Name     string
// 	Email    string
// 	Password string
// }

// func (u *User) ValidationSpec() Spec {
// 	return Spec{
// 		FieldSpec(u.Name, MaxLen(15), MinLen(12)),
// 		FieldSpec(u.Password, MaxLen(26), MinLen(8)),
// 		FieldSpec(u.Email, Email()),
// 	}
// }

// func adHoc() {
// 	var u User
// 	Validate(u, func(u User) Spec {
// 		return Spec{
// 			FieldSpec(u.Name, MaxLen(15)),
// 		}
// 	})

// }
