package structs

// ConfigElement represents a single configuration element.
type ConfigElement struct {
	ID        string
	Checksum  string
	Type      ConfigType
	Value     *string
	WriteOnly bool
}
