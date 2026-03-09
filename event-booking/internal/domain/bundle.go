package domain

type Bundle struct {
	ID             int64
	EventID        int64
	Code           string
	SectorName     string
	BundleBuyCount int64
	BundleGetCount int64
}
