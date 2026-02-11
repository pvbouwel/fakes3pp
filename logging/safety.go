package logging

import "log/slog"

type SafeStringerOptions struct {
	//Max lenght the string can have if it exceeds this it gets truncated and last characters will be ...truncated
	maxLength int
}

const truncatedSuffix = "...(truncated)"

func WithMaxLenght(maxLength int) func(*SafeStringerOptions) {
	return func(s *SafeStringerOptions) {
		s.maxLength = maxLength
	}
}

func (ss SafeStringerOptions) makeSafe(s string) string {
	if len(s) > ss.maxLength {
		s = s[:ss.maxLength-len(truncatedSuffix)] + truncatedSuffix
	}
	return s
}

func SafeString(s string, optFns ...func(*SafeStringerOptions)) slog.Value {
	//Default options used
	var opts = SafeStringerOptions{
		maxLength: 16 * 1024,
	}
	for _, optFn := range optFns {
		optFn(&opts)
	}
	return slog.StringValue(opts.makeSafe(s))
}
