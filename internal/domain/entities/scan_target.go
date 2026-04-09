package entities

type ScanTarget struct {
	RepositoryURL string
	ImageRef      string
	TargetURL     string
}

func (t ScanTarget) Primary() string {
	switch {
	case t.RepositoryURL != "":
		return t.RepositoryURL
	case t.ImageRef != "":
		return t.ImageRef
	default:
		return t.TargetURL
	}
}
