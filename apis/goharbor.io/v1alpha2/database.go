package v1alpha2

//go:generate stringer -type=ComponentWithDatabase -linecomment
type ComponentWithDatabase int

const (
	CoreDatabase         ComponentWithDatabase = iota // core
	NotaryServerDatabase                              // notaryserver
	NotarySignerDatabase                              // notarysigner
	ClairDatabase                                     // clair
)

func (r ComponentWithDatabase) DBName() string {
	return r.String()
}
