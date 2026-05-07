package browserauth

import "github.com/major/volumeleaders-go/volumeleaders"

// Option configures FindSession behavior.
type Option interface {
	apply(*findConfig)
}

type findConfig struct {
	browser        string
	profile        string
	skipValidation bool
	clientOpts     []volumeleaders.Option
}

type optionFunc func(*findConfig)

func (f optionFunc) apply(cfg *findConfig) {
	f(cfg)
}

// WithBrowser restricts session discovery to cookies from browser.
func WithBrowser(browser string) Option {
	return optionFunc(func(cfg *findConfig) {
		cfg.browser = browser
	})
}

// WithProfile restricts session discovery to cookies from profile.
func WithProfile(profile string) Option {
	return optionFunc(func(cfg *findConfig) {
		cfg.profile = profile
	})
}

// WithoutValidation skips fetching an XSRF token after cookie discovery.
func WithoutValidation() Option {
	return optionFunc(func(cfg *findConfig) {
		cfg.skipValidation = true
	})
}

// WithClientOptions appends VolumeLeaders client options used for validation.
func WithClientOptions(opts ...volumeleaders.Option) Option {
	return optionFunc(func(cfg *findConfig) {
		cfg.clientOpts = append(cfg.clientOpts, opts...)
	})
}
