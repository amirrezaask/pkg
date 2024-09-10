# Doctor Golang Packages
# Features
## AMQP (Rabbit)
## Consumer
## Publisher
## API

## Sequel
All applications that we write needs at least one database. `sequel` package makes it easy to use databases without sacrificing any performance and also provides great tools for visibility and monitoring, As an extension to golang standard "database/sql" package, we extend functionalities by using golang built-in infrastructures to provide a simpler and more conventional way of dealing with relational databases. 
#### How to start
To start using a relational database with `sequel` package you just need to create a connection as you would do when using `database/sql`, and actually you don't even need this step if you just want to test stuff but we recommend you use our constructor because we add metrics on top default `database/sql`.
```go
	db, err := New(DataSource{
		Driver:                "sqlite3",
		Name:                  "testingconnection",
		ConnectionString:      "file::memory:?mode=memory&cache=shared",
		MetricsNamespace:      "createdat",
		MaxOpenConnections:    50,
		MaxIdleConnections:    50,
		IdleConnectionTimeout: 10 * time.Second,
		OpenConnectionTimeout: 10 * time.Second,
	})
	if err != nil {
		panic(err)
	}
```
#### Models
sequel is built on top of an interface called `Model` which as you might guess is just an abstraction for all database records and this interface is a simple one-method interface:
```go
type Model interface{
	Schema() *Schema
}
```
But what is the `Schema` type ?
Schema is how you want to describe your type inside your database and also mapping between database columns and your type fields, let's demonstrate this with an example:
```go
type User struct {
	ID        int64
	Name      string
	LastName  string
	Credit    int
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (u *User) SequelRecordSpec() RecordSpec {
	return RecordSpec{
		Connection: "testingconnection",
		Table:      "users",
		Columns: []Column{
			{Name: "id", Ptr: &u.ID},
			{Name: "name", Ptr: &u.Name},
			{Name: "last_name", Ptr: &u.LastName},
			{Name: "credit", Ptr: &u.Credit},
			{Name: "created_at", Ptr: &u.CreatedAt},
			{Name: "updated_at", Ptr: &u.UpdatedAt},
		},
	}
}
```
By just these few lines you are telling database package what is your table name for this type, what columns are there and how they are mapped to struct fields and which pointer to use for which column when scanning query results (remember that we don't use reflection at all so all information should be available to us at compile time.)
Now that we have a model let's do a simple CRUD for it.
```go


func main() {
	db, err := New(DataSource{
		Driver:                "sqlite3",
		Name:                  "testingconnection",
		ConnectionString:      "file::memory:?mode=memory&cache=shared",
		MetricsNamespace:      "createdat",
		MaxOpenConnections:    50,
		MaxIdleConnections:    50,
		IdleConnectionTimeout: 10 * time.Second,
		OpenConnectionTimeout: 10 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	command := createMigrationCommand(sqlite, &User{})
	_, err = db.Exec(command)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		Insert(&User{
			Name:     uuid.NewString(),
			LastName: uuid.NewString(),
			Credit:   88,
		})
	}
	user := User{
		Name:     "Amirreza",
		LastName: "Ask",
		Credit:   999999,
	}
	_, err = Insert(&user)
	if err != nil {
		panic(err)
	}

	fmt.Println("Amirreza ID is", user.ID)

	// Now let's do an update
	user.Credit += 1000
	err = Save(&user)
	if err != nil {
		panic(err)
	}

	// And now let's delete it.
	_, err = Delete(&user)
	if err != nil {
		panic(err)
	}

	// Now let us get a list of all our users with almost no money and give them a promotion.
	poorUsers, err := Query[User]("SELECT * FROM users WHERE credit < 100")
	if err != nil {
		panic(err)
	}
	for _, user := range poorUsers {
		// dump.This(user)
		user.Credit += 1000000
	}
	// and now let's do a bulk update on our sequel.
	err = Save(poorUsers...)
	if err != nil {
		panic(err)
	}

	tableInfo, err := ToMap(db.Query("PRAGMA table_info(users)"))
	if err != nil {
		panic(err)
	}
	for range tableInfo {
		// world is your oyster
	}

	// We highly recommend that you don't use ToMap in your production code since it's not memory efficient, but for scenarios
	// that you just want to see what a query returns and you don't feel like creating a model it's really handy.
}
```
Summary of exported symbols from `sequel` package.
```go
// Main interface extracted from `database/sql`.DB type to let users pass either that or our connection object.
type Interface interface{ ... } 
    func Debug(i Interface) Interface
    func New(ds DataSource) (Interface, error)
func Delete(m Record) (sql.Result, error)
func Insert(obj Record) (sql.Result, error)
func Query[M any, T pointer[M]](q string, args ...any) ([]T, error)
func Save[T Record](objs ...T) error
func Scan[M any, T pointer[M]](rows *sql.Rows, err error) ([]T, error)
func ToMap(rows *sql.Rows, err error) ([]map[string]interface{}, error)
type Record interface{ ... } // Main interface that structs can implement to represent 
type RecordSpec struct{ ... }
func AssertDb(t *testing.T, db Interface, tables ...string) *dbAssertions // used to assert database state in tests.
type AfterReadHook func(m Record) error
type BeforeWriteHook func(m Record) error
type Column struct{ ... }
type ConnectionOptions struct{ ... }
type DataSource struct{ ... }
type MockDb struct{ ... }
    func NewMockDb(t *testing.T, connectionName string, target *Interface, models ...Record) *MockDb
```

## kv
## Dump
## Errors
## HttpClient
## Logging
## Must
## ObjectStore
## Retry
## Set
## Test
## Tracing


# Usage
## Put following credentials in $HOME/.netrc
```
machine gitlab.snappcloud.io 
login .
password glpat-FgvkiH-xd-P4HuB4vTZC
```

## Set GOPRIVATE env
```
export GOPRIVATE='gitlab.snappcloud.io'
```
