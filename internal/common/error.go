package common

import "fmt"

func WrapErrorf(err error, format string, a ...any) error {
	wrapFormat := fmt.Sprintf("%s: %%w", format)
	return fmt.Errorf(wrapFormat, append(a, err)...)
}
