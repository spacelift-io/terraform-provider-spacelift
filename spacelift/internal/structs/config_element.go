package structs

// ConfigElement represents a single configuration element.
type ConfigElement struct {
	ID        string
	Checksum  string
	FileMode  *string
	Type      ConfigType
	Value     *string
	WriteOnly bool
}
