package sbom

import (
	"encoding/xml"
	"github.com/google/uuid"
	"github.com/package-url/packageurl-go"
	"iq-scm-audit/github"
	"strings"
)

type Sbom struct {
	XMLName xml.Name `xml:"bom"`
	XMLNs string `xml:"xmlns,attr"`
	Version string `xml:"version,attr"`
	SerialNumber string `xml:"serialNumber,attr"`
	Components Components `xml:"components"`
}

type Components struct {
	Component[] Component
}

type Component struct {
	XMLName xml.Name `xml:"component"`
	Type string `xml:"type,attr"`
	Group string `xml:"group"`
	Name string `xml:"name"`
	Version string `xml:"version"`
	Purl string `xml:"purl"`
}

func NewSbom(dependencies[] github.Dependency) *Sbom {
	sbom := new(Sbom)
	sbom.XMLNs = "http://cyclonedx.org/schema/bom/1.1"
	sbom.Version = "1"
	sbom.SerialNumber = "urn:uuid:" + uuid.New().String()

	for _, dependency := range dependencies {
		if len(dependency.Requirements) > 2 {
			component := new(Component)
			component.Type = "library"

			v := dependency.Requirements[2:]
			lowerPackageManager := strings.ToLower(dependency.PackageManager)
			switch lowerPackageManager {
			case "maven":
				ga := strings.Split(dependency.PackageName, ":")
				component.Group = ga[0]
				component.Name = ga[1]
				component.Version = v
				qualifier := packageurl.QualifiersFromMap(map[string] string{
					"type": "jar",
				})
				component.Purl = packageurl.NewPackageURL("maven", ga[0], ga[1], v, qualifier, "").String()
			case "npm", "nuget":
				component.Name = dependency.PackageName
				component.Version = v
				component.Purl = packageurl.NewPackageURL(lowerPackageManager, "", dependency.PackageName, v, nil, "").String()
			}

			sbom.Components.Component = append(sbom.Components.Component, *component)
		}
	}

	return sbom
}