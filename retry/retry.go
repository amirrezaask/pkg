package retry

import (
	"fmt"
	"time"
)

func Do(f func() error, retries int, backoff time.Duration) error {
	err := f()
	if err != nil {
		for i := 0; i < retries; i++ {
			time.Sleep(backoff)
			err = f()
			if err != nil {
				continue
			}
			break

		}
	}

	if err != nil {
		return fmt.Errorf("retried for %d times: %w", retries, err)
	}
	return nil
}
