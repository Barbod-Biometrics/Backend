package exception

import "fmt"

type NotFoundError struct {
	Item string
}

func (nf NotFoundError) Error() string {
	return fmt.Sprintf("item: %s not found.", nf.Item)
}
