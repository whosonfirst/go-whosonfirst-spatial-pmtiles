package placetypes

import (
	"log"
	"sync"
)

type WOFPlacetypeName struct {
	// Lang is the RFC 5646 (BCP-47) language tag for the placetype name
	Lang string `json:"language"`
	Kind string `json:"kind"`
	// Name is the name of the placetype (in the language defined by `Lang`)
	Name string `json:"name"`
}

type WOFPlacetypeAltNames map[string][]string

// Type WOFPlacetype defines an individual placetype encoded in a `WOFPlacetypeSpecification`
// instance. The choice of naming this "WOFPlacetype" is unfortunate because since it is easily
// confused with the actual JSON definition files for placetypes. However, we're stuck with it
// for now in order to preserve backwards compatibility. Womp womp...
//
// This needs to be renamed to "Placetype" or something either at a /v1 or a v2 release. Either
// way it will be a breaking change. Doing it v2 (even if there is no explicit v1 release)
// might be "cleaner"...
type WOFPlacetype struct {
	Id     int64   `json:"id"`
	Name   string  `json:"name"`
	Role   string  `json:"role"`
	Parent []int64 `json:"parent"`
	// AltNames []WOFPlacetypeAltNames		`json:"names"`
}

// String returns the value of the `Name` property for 'pt'.
func (pt *WOFPlacetype) String() string {
	return pt.Name
}

// IsCorePlacetype returns a boolean value if 'pt' is one of the "core" Who's On First placetypes.
func (pt *WOFPlacetype) IsCorePlacetype() bool {
	return isCorePlacetype(pt.Name)
}

// What follows is legacy code. Specifically this is code that was developed before there was
// the notion of multiple placetype specifications that would be merged. For example the "core"
// Who's On First placetype specification and the SFO Museum placetype specification defined in
// sfomuseum/go-sfomuseum-placetypes and sfomuseum/sfomuseum-placetypes packages. As such there
// was no need to expose the underlying `WOFPlacetypeSpecification` instance and all methods
// were assumed to operate on the internal specification instance. Subsequently all the code that
// used to be defined as standalone methods without reference to any specific placetype specification
// has been moved in specification.go (and are now methods on individual WOFPlacetypeSpecification
// instances). These methods have been preserved for backwards compatibility.

var specification *WOFPlacetypeSpecification
var core_placetypes *sync.Map

func init() {

	s, err := DefaultWOFPlacetypeSpecification()

	if err != nil {
		log.Fatalf("Failed to load default WOF specification, %v", err)
	}

	all_placetypes, err := s.Placetypes()

	if err != nil {
		log.Fatalf("Failed to derive placetypes from spec, %v", err)
	}

	core_placetypes = new(sync.Map)

	for _, pt := range all_placetypes {
		core_placetypes.Store(pt.Name, pt.Id)
	}

	specification = s
}

// GetPlacetypesByName returns the `WOFPlacetype` instance associated with 'name'.
func GetPlacetypeByName(name string) (*WOFPlacetype, error) {
	return specification.GetPlacetypeByName(name)
}

// GetPlacetypesByName returns the `WOFPlacetype` instance associated with 'id'.
func GetPlacetypeById(id int64) (*WOFPlacetype, error) {
	return specification.GetPlacetypeById(id)
}

// AppendPlacetype appends 'pt' to the catalog of available placetypes.
func AppendPlacetype(pt WOFPlacetype) error {
	return specification.AppendPlacetype(pt)
}

// AppendPlacetypeSpecification appends the placetypes defined in 'other_spec' to the catalog of available placetypes in 'spec'.
func AppendPlacetypeSpecification(spec *WOFPlacetypeSpecification) error {
	return specification.AppendPlacetypeSpecification(spec)
}

// Placetypes returns all the known placetypes which are descendants of "planet" for the 'common', 'optional', 'common_optional', and 'custom' roles.
func Placetypes() ([]*WOFPlacetype, error) {

	roles := []string{
		COMMON_ROLE,
		OPTIONAL_ROLE,
		COMMON_OPTIONAL_ROLE,
		CUSTOM_ROLE,
	}

	return PlacetypesForRoles(roles)
}

// Placetypes returns all the known placetypes which are descendants of "planet" whose role match any of those defined in 'roles'.
func PlacetypesForRoles(roles []string) ([]*WOFPlacetype, error) {
	return specification.PlacetypesForRoles(roles)
}

// IsValidPlacetypeId returns a boolean value indicating whether 'name' is a known and valid placetype name.
func IsValidPlacetype(name string) bool {
	return specification.IsValidPlacetype(name)
}

// IsValidPlacetypeId returns a boolean value indicating whether 'id' is a known and valid placetype ID.
func IsValidPlacetypeId(id int64) bool {
	return specification.IsValidPlacetypeId(id)
}

// Returns true is 'b' is an ancestor of 'a'.
func IsAncestor(a *WOFPlacetype, b *WOFPlacetype) bool {
	return specification.IsAncestor(a, b)
}

// Returns true is 'b' is a descendant of 'a'.
func IsDescendant(a *WOFPlacetype, b *WOFPlacetype) bool {
	return specification.IsDescendant(a, b)
}

// Children returns the immediate child placetype of 'pt'.
func Children(pt *WOFPlacetype) []*WOFPlacetype {
	return specification.Children(pt)
}

// Descendants returns the descendants of role "common" for 'pt'.
func Descendants(pt *WOFPlacetype) []*WOFPlacetype {
	return specification.Descendants(pt)
}

// DescendantsForRoles returns the descendants matching any role in 'roles' for 'pt'.
func DescendantsForRoles(pt *WOFPlacetype, roles []string) []*WOFPlacetype {
	return specification.DescendantsForRoles(pt, roles)
}

// Ancestors returns the ancestors of role "common" for 'pt'.
func Ancestors(pt *WOFPlacetype) []*WOFPlacetype {
	return AncestorsForRoles(pt, []string{"common"})
}

// AncestorsForRoles returns the ancestors matching any role in 'roles' for 'pt'.
func AncestorsForRoles(pt *WOFPlacetype, roles []string) []*WOFPlacetype {
	return specification.AncestorsForRoles(pt, roles)
}

func isCorePlacetype(n string) bool {
	_, exists := core_placetypes.Load(n)
	return exists
}
