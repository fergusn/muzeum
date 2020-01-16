package model

// Package structure from purl
type Package struct {
	Type       string // The package "type" or package "protocol" such as maven, npm, nuget, gem, pypi, etc
	Namespace  string // some name prefix such as a Maven groupid, a Docker image owner, a GitHub user or organization
	Name       string
	Version    string
	Qualifiers map[string]string
	Subpath    string
}
