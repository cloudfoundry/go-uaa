package uaa

import (
	"fmt"
)

// Config is used to access the UAA API.
type Config struct {
	Verbose          bool
	ZoneSubdomain    string
	Targets          map[string]Target
	ActiveTargetName string
}

// Target is a UAA endpoint.
type Target struct {
	BaseURL           string
	SkipSSLValidation bool
	Contexts          map[string]AuthContext
	ActiveContextName string
}

// AuthContext is a container for the token used to access UAA.
type AuthContext struct {
	ClientID  string    `json:"client_id"`
	GrantType GrantType `json:"grant_type"`
	Username  string    `json:"username"`
	TokenResponse
}

// NewConfigWithServerURL creates a new config with the given URL.
func NewConfigWithServerURL(url string) Config {
	c := NewConfig()
	t := NewTarget()
	t.BaseURL = url
	c.AddTarget(t)
	return c
}

// NewContextWithToken creates a new config with the given token.
func NewContextWithToken(accessToken string) AuthContext {
	ctx := AuthContext{}
	ctx.AccessToken = accessToken
	return ctx
}

// NewConfig creates a config that is initialized with an empty map of targets.
func NewConfig() Config {
	c := Config{}
	c.Targets = map[string]Target{}
	return c
}

// NewTarget creates a target that is initialized with an empty map of contexts.
func NewTarget() Target {
	t := Target{}
	t.Contexts = map[string]AuthContext{}
	return t
}

// AddTarget adds the given target to the config, and sets the active target to
// the given target.
func (c *Config) AddTarget(newTarget Target) {
	c.Targets[newTarget.name()] = newTarget
	c.ActiveTargetName = newTarget.name()
}

// AddContext adds the given context to the active target.
func (c *Config) AddContext(newContext AuthContext) {
	if c.Targets == nil {
		c.Targets = map[string]Target{}
	}
	t := c.Targets[c.ActiveTargetName]
	if t.Contexts == nil {
		t.Contexts = map[string]AuthContext{}
	}
	t.Contexts[newContext.name()] = newContext
	t.ActiveContextName = newContext.name()
	c.AddTarget(t)
}

// GetActiveTarget gets the active target.
func (c Config) GetActiveTarget() Target {
	return c.Targets[c.ActiveTargetName]
}

// GetActiveContext gets the active context.
func (c Config) GetActiveContext() AuthContext {
	return c.GetActiveTarget().GetActiveContext()
}

// GetActiveContext gets the active context.
func (t Target) GetActiveContext() AuthContext {
	return t.Contexts[t.ActiveContextName]
}

func (t Target) name() string {
	return "url:" + t.BaseURL
}

func (uc AuthContext) name() string {
	return fmt.Sprintf("client:%v user:%v grant_type:%v", uc.ClientID, uc.Username, uc.GrantType)
}
