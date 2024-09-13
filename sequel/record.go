package sequel

import (
	"fmt"
	"github.com/amirrezaask/pkg/errors"
	"reflect"
)

type Record interface {
	SequelRecordSpec() RecordSpec
}

type Column struct {
	Name    string
	Ptr     any
	Options int
}

type BeforeWriteHook func(m Record) error
type AfterReadHook func(m Record) error

type RecordSpec struct {
	Connection  string
	Table       string
	Columns     []Column
	BeforeWrite []BeforeWriteHook
	AfterRead   []AfterReadHook
}

type recordSpec struct {
	connectionName string
	pk             *int64
	pkName         string
	createdAtName  string
	updatedAtName  string
	columnOptions  map[string]int
	fillable       []string
	table          string
	valueMap       map[string]any
	beforeWrite    []BeforeWriteHook
	afterRead      []AfterReadHook
}

const (
	_ = 1 << iota
	PK
	NoColumn
	CreatedAt
	UpdatedAt
)

type columnSpec struct {
	Name string
	Type reflect.Type
	IsPK bool
}

func (s *recordSpec) GetColumns() (string, []columnSpec) {
	var specs []columnSpec
	for _, fil := range s.fillable {
		var isPK bool
		if s.columnOptions[fil]&PK == PK {
			isPK = true
		}
		specs = append(specs, columnSpec{
			Name: fil,
			Type: reflect.TypeOf(s.valueMap[fil]).Elem(),
			IsPK: isPK,
		})
	}

	return s.table, specs
}

func (r RecordSpec) intoInternalRepr() (*recordSpec, error) {
	internal := &recordSpec{
		connectionName: r.Connection,
		table:          r.Table,
		valueMap:       map[string]any{},
	}

	for i, colSpec := range r.Columns {
		if colSpec.Name == "" {
			return nil, fmt.Errorf("invalid RecordSpec for table %s, no column name was set for column %d", r.Table, i)
		}
		if colSpec.Ptr == nil {
			return nil, fmt.Errorf("invalid RecordSpec for table %s, nil pointer was set for %s column", r.Table, colSpec.Name)
		}
		internal.valueMap[colSpec.Name] = colSpec.Ptr
		if colSpec.Options != 0 {
			if colSpec.Options&PK == PK {
				internal.pk = colSpec.Ptr.(*int64)
				internal.pkName = colSpec.Name
			} else if colSpec.Options&CreatedAt == CreatedAt {
				internal.createdAtName = colSpec.Name
			} else if colSpec.Options&UpdatedAt == UpdatedAt {
				internal.updatedAtName = colSpec.Name
			}

			if (colSpec.Options&PK != PK) && (colSpec.Options&NoColumn != NoColumn) {
				internal.fillable = append(internal.fillable, colSpec.Name)
			}
		} else {
			internal.fillable = append(internal.fillable, colSpec.Name)

		}
	}

	if internal.createdAtName == "" {
		if internal.valueMap["created_at"] != nil {
			internal.createdAtName = "created_at"
		}
	}

	if internal.updatedAtName == "" {
		if internal.valueMap["updated_at"] != nil {
			internal.updatedAtName = "updated_at"
		}
	}

	if internal.pk == nil {
		if internal.valueMap["id"] != nil {
			if _, isIntPtr := internal.valueMap["id"].(*int64); isIntPtr {
				internal.pk = internal.valueMap["id"].(*int64)
			}
		}
		idIndex := 0
		for i, fil := range internal.fillable {
			if fil == "id" {
				idIndex = i
			}
		}
		internal.pkName = "id"

		internal.fillable = append(internal.fillable[:idIndex], internal.fillable[idIndex+1:]...)
	}

	internal.beforeWrite = r.BeforeWrite
	internal.afterRead = r.AfterRead

	return internal, nil
}

func (r RecordSpec) Validate() error {
	s, err := r.intoInternalRepr()
	if err != nil {
		return fmt.Errorf("cannot validate record spec for table '%s': %w", r.Table, err)
	}
	if s.table == "" {
		return errors.Newf("No table has been defined")
	}
	if len(s.fillable) < 1 {
		return errors.Newf("No columns defined for model of table %s", s.table)
	}
	if s.pk == nil || fmt.Sprintf("%T", s.pk) != "*int64" {
		return errors.Newf("primary key must be defined and it's type should be *int64, table %s has no valid primary key", s.table)
	}

	if len(s.valueMap) == 0 {
		return errors.Newf("no field mapping defined for model of table %s", s.table)
	}
	//TODO(amirreza): optional reflection check for validity of types.

	return nil
}
