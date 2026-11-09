package ascii

type Config struct {
	values map[string]any
}

func NewConfig() *Config {
	return &Config{
		values: make(map[string]any),
	}
}

func (c *Config) Set(key string, value any) {
	c.values[key] = value
}

func (c *Config) GetString(key string, defaultValue string) string {
	val, ok := c.values[key]
	if !ok {
		return defaultValue
	}
	if str, okType := val.(string); okType {
		return str
	}
	return defaultValue
}

type Option func(*Config)

func WithStyle(style string) Option {
	return func(c *Config) {
		c.Set("style", style)
	}
}
