package contracts

type ProbeConfig struct {
	Name        string
	TypeKey     string
	ProviderKey string
	Status      string
	Metadata    map[string]any
	Plaintext   string
	HasSecret   bool
}
