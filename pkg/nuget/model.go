package nuget

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	purl "github.com/package-url/packageurl-go"
)

type ResourceType string
type VersionRange string

const (
	PackagePublish       ResourceType = "PackagePublish/2.0.0"
	SearchQueryService   ResourceType = "SearchQueryService"
	RegistrationsBaseUrl ResourceType = "RegistrationsBaseUrl"
	PackageBaseAddress   ResourceType = "PackageBaseAddress/3.0.0"
)

type ServiceIndex struct {
	Version   string     `json:"version"`
	Resources []Resource `json:"resources"`
}
type Resource struct {
	ID      string       `json:"@id"`
	Type    ResourceType `json:"@type"`
	Comment string       `json:"comment"`
}

type Package struct {
	XMLName  xml.Name `xml:"package"`
	Metadata Metadata `xml:"metadata"`
	Files    []File   `xml:"files"`
}

type Nuspec struct {
	Package Package `xml:"package"`
}

func (pkg *Package) PURL() purl.PackageURL {
	return purl.PackageURL{
		Type:    "nuget",
		Name:    strings.ToLower(pkg.Metadata.ID),
		Version: strings.ToLower(pkg.Metadata.Version),
	}
}

type version struct {
	major  int
	minor  int
	path   int
	suffix string
}

var (
	vs = `(?P<major>\d+)\.(?P<minor>\d+)(\.(?P<patch>\d+)){0,1}(\-(?P<suffix>[a-z0-9]+)){0,1}`

	min      = regexp.MustCompile(`^` + vs + `$`)
	exact    = regexp.MustCompile(`^\[` + vs + `\]$`)
	minmax   = regexp.MustCompile(`^[\[\(]` + vs + `,` + vs + `[\]\)]$`)
	wildcard = regexp.MustCompile(`^$`)
)

type Metadata struct {
	ID          string `xml:"id"`
	Version     string `xml:"version"`
	Description string `xml:"description"`
	Authors     string `xml:"authors"`

	Title                    string        `xml:"title"`
	Owners                   string        `xml:"owners"`
	ProjectURL               string        `xml:"projectUrl"`
	LicenseURL               string        `xml:"licenseUrl"`
	License                  string        `xml:"license"`
	IconURL                  string        `xml:"iconUrl"`
	RequireLicenseAcceptance bool          `xml:"requireLicenseAcceptance"`
	DevelopmentDependency    bool          `xml:"developmentDependency"`
	Summary                  string        `xml:"summary"`
	ReleaseNotes             string        `xml:"releaseNotes"`
	Copyright                string        `xml:"copyright"`
	Language                 string        `xml:"language"`
	Tags                     string        `xml:"tags"`
	Serviceable              string        `xml:"serviceable"`
	Dependencies             *Dependencies `xml:"dependencies"`
}

type PackageType struct {
	Name    string `xml:"name,attr"`
	Version string `xml:"name,attr"`
}

type Any struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

type Dependencies struct {
	Groups       []*Group
	Dependencies []*Dependency
}

func (deps *Dependencies) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	xs := &deps.Dependencies
	for {
		n, _ := d.Token()
		if s, ok := n.(xml.StartElement); ok {
			if s.Name.Local == "group" {
				grp := &Group{
					TargetFramework: s.Attr[0].Value,
					Dependencies:    []*Dependency{},
				}
				deps.Groups = append(deps.Groups, grp)
				xs = &grp.Dependencies
			} else if s.Name.Local == "dependency" {
				dep := Dependency{}
				err := d.DecodeElement(&dep, &s)
				*xs = append(*xs, &dep)
				fmt.Println(err)
			}
		} else if e, ok := n.(xml.EndElement); ok {
			if e.Name.Local == "dependencies" {
				break
			}
		}
	}
	return nil
}

type Dependency struct {
	XMLName xml.Name     `xml:"dependency"`
	ID      string       `xml:"id,attr"`
	Version VersionRange `xml:"version,attr"`
	Include string       `xml:"include,attr"`
	Exclude string       `xml:"exclude,attr"`
}

type Group struct {
	TargetFramework string
	Dependencies    []*Dependency
}

type FrameworkAssembly struct {
	AssemblyName    string `xml:"assemblyName,attr"`
	TargetFramework string `xml:"targetFramework,attr,omitEmpty"`
}
type Reference struct {
	File string
}

type ReferenceGroup struct {
	TargetFramework string      `xml:"targetFramework,attr"`
	References      []Reference `xml:-`
}

type ContentFile struct {
	Include      string `xml:"include,attr"`
	Exclude      string `xml:"excluse,attr"`
	BuildAction  string `xml:"buildAction,attr"`
	CopyToOutput bool   `xml:"copyToOutput,attr"`
	Flatten      bool   `xml:"flatten,attr"`
}

type File struct {
	Src     string `xml:"src,attr"`
	Target  string `xml:"target,attr"`
	Exclude string `xml:"exclude,attr"`
}

func (v VersionRange) Min() string {
	if v[0] == '(' {

	}
	if v[0] == '[' {

	}
	return string(v)
}
