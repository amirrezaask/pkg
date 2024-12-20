package ab

type Interface interface {
	IsUserEligible(feature string, userID int64) bool
}
