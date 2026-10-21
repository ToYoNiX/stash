// Package app holds the use cases that span more than one feature.
// Feature packages (containers, items) never import each other;
// anything that needs both of them lives here.
package app

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ToYoNiX/stash/internal/containers"
	"github.com/ToYoNiX/stash/internal/items"
)

// AddItem stores an item after making sure its container exists.
// The foreign key in the database is the real guarantee; this check
// is here to give the caller a clear error instead of a constraint
// failure.
func AddItem(db *sql.DB, i *items.Item) (int, error) {
	if i.ContainerId == nil {
		return 0, errors.New("Parent container not provided")
	}

	exists, err := containers.DoesExist(db, *i.ContainerId)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("Container %d not found", *i.ContainerId)
	}

	return items.Insert(db, i)
}
