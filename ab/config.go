package ab

import (
	"fmt"

	"github.com/amirrezaask/go-std/set"
)

type ABConfig struct {
	Whitelist  set.Set[string]
	BucketSize int64
}

func (a *ABConfig) IsUserEligible(userID int64) bool {
	return a.Whitelist.Exists(fmt.Sprint(userID)) || (userID%100) < a.BucketSize
}
