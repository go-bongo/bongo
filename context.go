package bongo

// Context struct
type Context struct {
	set map[string]interface{}
}

// Get ...
func (c *Context) Get(key string) interface{} {
	if value, ok := c.set[key]; ok {
		return value
	}
	return nil
}

func (c *Context) Delete(key string) bool {
	if _, ok := c.set[key]; ok {
		delete(c.set, key)
		return true
	}
	return false
}

// Set ...
func (c *Context) Set(key string, value interface{}) {
	if c.set == nil {
		c.set = make(map[string]interface{})
	}
	c.set[key] = value
}
