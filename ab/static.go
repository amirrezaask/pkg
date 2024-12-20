package ab

import (
	"fmt"

	"github.com/amirrezaask/pkg/set"
)

type Static struct {
	Whitelist  set.Set[string]
	BucketSize int64
}

func (a *Static) IsUserEligible(feature string, userID int64) bool {
	return a.Whitelist.Exists(fmt.Sprint(userID)) || (userID%100) < a.BucketSize
}
