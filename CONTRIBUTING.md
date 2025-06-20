# Contributing

## Developing

Please install the `pre-commit` to enforce the code conventions and alignment.

```shell
pip install pre-commit
```

Install the required pre-commit hooks.

```shell
pre-commit install -t commit-msg
```

## Writing unit tests

### Sqlmock issues

When writing repository unit tests we started relying on sqlmock to mock the db driver and unit test SQL queries etc.

Sqlmock seems to have a few issues that prevents it to work correctly and consistently unless we do certain things.
Writing here to make sure we have a small guideline to write unit tests the way they seems to work, when working with
sqlmock.

Provide a common setup logic for uni test leveraging Go standard library `TestMain(m *testing.M)` function.

Example:

```go
package test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var (
	db      *sql.DB
	mock    sqlmock.Sqlmock
	mockErr error
)

func TestMain(m *testing.M) {
	db, mock, mockErr = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if mockErr != nil {
		panic(mockErr)
	}
	defer db.Close()

	m.Run()

}

```

When writing a unit test, make sure to include this call to allow the sqlmock to function properly

```go
if err := mock.ExpectationsWereMet(); err != nil {
    t.Fatalf("there were unfulfilled expectations: %v", err)
}

```

This seems to also "reset" the mock, but documentation is not clear about it.

