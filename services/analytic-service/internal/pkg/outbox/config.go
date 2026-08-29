package outbox

type Config struct {
	LockedKeysLimit       int `yaml:"locked_keys_limit"`
	MessagesLimit         int `yaml:"messages_limit"`
	MaxErrCountForMessage int `yaml:"max_err_count_for_message"`
}
