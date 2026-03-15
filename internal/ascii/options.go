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

func (c *Config) GetInt(key string, defaultValue int) int {
	val, ok := c.values[key]
	if !ok {
		return defaultValue
	}
	if i, okType := val.(int); okType {
		return i
	}
	return defaultValue
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	val, ok := c.values[key]
	if !ok {
		return defaultValue
	}
	if b, okType := val.(bool); okType {
		return b
	}
	return defaultValue
}

type Option func(*Config)

func WithStyle(style string) Option {
	return func(c *Config) {
		c.Set("style", style)
	}
}

func WithWidth(width int) Option {
	return func(c *Config) {
		c.Set("width", width)
	}
}

func WithHeight(height int) Option {
	return func(c *Config) {
		c.Set("height", height)
	}
}

func WithColor(useColor bool) Option {
	return func(c *Config) {
		c.Set("useColor", useColor)
	}
}
