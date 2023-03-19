package pkg

import (
	"context"
	"fmt"
)

func CtxGetValue[OUT any](ctx context.Context, key string) (OUT, error) {
	out := ctx.Value(key)
	outCasted, ok := out.(OUT)
	if !ok {
		return outCasted, fmt.Errorf("expected %T found %T", outCasted, out)
	}

	return outCasted, nil
}
