package router

import (
	"regexp"
)

var (
	compactSlashRegexp = regexp.MustCompile(`/{2,}`)
	staticPathRegexp   = regexp.MustCompile(`^[^\\/{}*]+$`)
	wildcardPathRegexp = regexp.MustCompile(`^*[\da-zA-Z_\\-]*$`)
	paramPathRegexp    = regexp.MustCompile(`^{([a-zA-Z]\w*|_\w*[a-zA-Z\d]+\w*)}$`)
)

func Normalize(p string) string {
	p = compactSlashRegexp.ReplaceAllString(p, "/")
	if p == "" {
		return p
	}

	if p[0] == '/' {
		p = p[1:]
	}

	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}

func IsStatic(p string) bool {
	return staticPathRegexp.MatchString(p)
}

func IsWildcard(p string) bool {
	if p == "" {
		return false
	}
	return wildcardPathRegexp.MatchString(p)
}

func IsParam(p string) bool {
	if p == "" {
		return false
	}
	return paramPathRegexp.MatchString(p)
}
