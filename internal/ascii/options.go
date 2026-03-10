package ascii

type Config struct {
	values map[string]interface{}
}

func NewConfig() *Config {
	return &Config{
		values: make(map[string]interface{}),
	}
}

func (c *Config) Set(key string, value interface{}) {
	c.values[key] = value
}

func (c *Config) GetString(key string, defaultValue string) string {
	if val, ok := c.values[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (c *Config) GetInt(key string, defaultValue int) int {
	if val, ok := c.values[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultValue
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	if val, ok := c.values[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
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
