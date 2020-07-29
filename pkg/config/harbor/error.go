package harbor

import (
	"fmt"
)

type errNoConfigFound string

func (e errNoConfigFound) Error() string {
	return fmt.Sprintf("no config found for %s", string(e))
}
