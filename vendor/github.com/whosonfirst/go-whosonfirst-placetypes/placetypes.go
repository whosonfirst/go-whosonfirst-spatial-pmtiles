package placetypes

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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

type WOFPlacetype struct {
	Id     int64   `json:"id"`
	Name   string  `json:"name"`
	Role   string  `json:"role"`
	Parent []int64 `json:"parent"`
	// AltNames []WOFPlacetypeAltNames		`json:"names"`
}

var specification *WOFPlacetypeSpecification

var relationships *sync.Map

func init() {

	s, err := DefaultWOFPlacetypeSpecification()

	if err != nil {
		log.Fatal("Failed to load default WOF specification", err)
	}

	specification = s

	relationships = new(sync.Map)

	go func() {

		roles := []string{
			"common",
			"optional",
			"common_optional",
		}

		count_roles := len(roles)

		for i := 0; i < count_roles; i++ {

			pt_roles := roles[0:i]

			for _, pt := range specification.catalog {
				go Children(&pt)
				go DescendantsForRoles(&pt, pt_roles)
				go AncestorsForRoles(&pt, pt_roles)
			}
		}
	}()

}

func GetPlacetypeByName(name string) (*WOFPlacetype, error) {
	return specification.GetPlacetypeByName(name)
}

func GetPlacetypeById(id int64) (*WOFPlacetype, error) {
	return specification.GetPlacetypeById(id)
}

func AppendPlacetype(pt WOFPlacetype) error {
	return specification.AppendPlacetype(pt)
}

func AppendPlacetypeSpecification(spec *WOFPlacetypeSpecification) error {
	return specification.AppendPlacetypeSpecification(spec)
}

// Placetypes returns all the known placetypes for the 'common', 'optional' and 'common_optional' roles.
func Placetypes() ([]*WOFPlacetype, error) {

	roles := []string{
		"common",
		"optional",
		"common_optional",
	}

	return PlacetypesForRoles(roles)
}

func PlacetypesForRoles(roles []string) ([]*WOFPlacetype, error) {

	pl, err := GetPlacetypeByName("planet")

	if err != nil {
		return nil, fmt.Errorf("Failed to load 'planet' placetype, %w", err)
	}

	pt_list := DescendantsForRoles(pl, roles)

	pt_list = append([]*WOFPlacetype{pl}, pt_list...)
	return pt_list, nil
}

// IsValidPlacetypeId returns a boolean value indicating whether 'name' is a known and valid placetype name.
func IsValidPlacetype(name string) bool {

	for _, pt := range specification.Catalog() {

		if pt.Name == name {
			return true
		}
	}

	return false
}

// IsValidPlacetypeId returns a boolean value indicating whether 'id' is a known and valid placetype ID.
func IsValidPlacetypeId(id int64) bool {

	for str_id, _ := range specification.Catalog() {

		pt_id, err := strconv.Atoi(str_id)

		if err != nil {
			continue
		}

		pt_id64 := int64(pt_id)

		if pt_id64 == id {
			return true
		}
	}

	return false
}

// Returns true is 'b' is an ancestor of 'a'.
func IsAncestor(a *WOFPlacetype, b *WOFPlacetype) bool {

	roles := []string{
		"common",
		"optional",
		"common_optional",
	}

	str_roles := strings.Join(roles, "-")
	key := fmt.Sprintf("%d_%d_%s_is_ancestor", a.Id, b.Id, str_roles)

	v, ok := relationships.Load(key)

	if ok {
		return v.(bool)
	}

	is_ancestor := false

	for _, ancestor := range AncestorsForRoles(a, roles) {

		if ancestor.Name == b.Name {
			is_ancestor = true
			break
		}
	}

	relationships.Store(key, is_ancestor)
	return is_ancestor
}

// Returns true is 'b' is a descendant of 'a'.
func IsDescendant(a *WOFPlacetype, b *WOFPlacetype) bool {

	roles := []string{
		"common",
		"optional",
		"common_optional",
	}

	str_roles := strings.Join(roles, "-")
	key := fmt.Sprintf("%d_%d_%s_is_descendant", a.Id, b.Id, str_roles)

	v, ok := relationships.Load(key)

	if ok {
		return v.(bool)
	}

	is_descendant := false

	for _, descendant := range DescendantsForRoles(a, roles) {

		if descendant.Name == b.Name {
			is_descendant = true
			break
		}
	}

	relationships.Store(key, is_descendant)
	return is_descendant
}

// Children returns the immediate child placetype of 'pt'.
func Children(pt *WOFPlacetype) []*WOFPlacetype {

	key := fmt.Sprintf("%d_children", pt.Id)

	v, ok := relationships.Load(key)

	if ok {
		return v.([]*WOFPlacetype)
	}

	children := make([]*WOFPlacetype, 0)

	for _, details := range specification.Catalog() {

		for _, pid := range details.Parent {

			if pid == pt.Id {
				child_pt, _ := GetPlacetypeByName(details.Name)
				children = append(children, child_pt)
			}
		}
	}

	sorted := sortChildren(pt, children)

	relationships.Store(key, sorted)
	return sorted
}

func sortChildren(pt *WOFPlacetype, all []*WOFPlacetype) []*WOFPlacetype {

	kids := make([]*WOFPlacetype, 0)
	grandkids := make([]*WOFPlacetype, 0)

	for _, other := range all {

		is_grandkid := false

		for _, pid := range other.Parent {

			for _, p := range all {

				if pid == p.Id {
					is_grandkid = true
					break
				}
			}

			if is_grandkid {
				break
			}
		}

		if is_grandkid {
			grandkids = append(grandkids, other)
		} else {
			kids = append(kids, other)
		}
	}

	if len(grandkids) > 0 {
		grandkids = sortChildren(pt, grandkids)
	}

	for _, k := range grandkids {
		kids = append(kids, k)
	}

	return kids
}

// Descendants returns the descendants of role "common" for 'pt'.
func Descendants(pt *WOFPlacetype) []*WOFPlacetype {
	return DescendantsForRoles(pt, []string{"common"})
}

// DescendantsForRoles returns the descendants matching any role in 'roles' for 'pt'.
func DescendantsForRoles(pt *WOFPlacetype, roles []string) []*WOFPlacetype {

	str_roles := strings.Join(roles, "-")
	key := fmt.Sprintf("%d_descendants_%s", pt.Id, str_roles)

	v, ok := relationships.Load(key)

	if ok {
		return v.([]*WOFPlacetype)
	}

	descendants := make([]*WOFPlacetype, 0)
	descendants = fetchDescendants(pt, roles, descendants)

	relationships.Store(key, descendants)
	return descendants
}

func fetchDescendants(pt *WOFPlacetype, roles []string, descendants []*WOFPlacetype) []*WOFPlacetype {

	grandkids := make([]*WOFPlacetype, 0)

	for _, kid := range Children(pt) {

		descendants = appendPlacetype(kid, roles, descendants)

		for _, grandkid := range Children(kid) {
			grandkids = appendPlacetype(grandkid, roles, grandkids)
		}
	}

	for _, k := range grandkids {
		descendants = appendPlacetype(k, roles, descendants)
		descendants = fetchDescendants(k, roles, descendants)
	}

	return descendants
}

func appendPlacetype(pt *WOFPlacetype, roles []string, others []*WOFPlacetype) []*WOFPlacetype {

	do_append := true

	for _, o := range others {

		if pt.Id == o.Id {
			do_append = false
			break
		}
	}

	if !do_append {
		return others
	}

	has_role := false

	for _, r := range roles {

		if pt.Role == r {
			has_role = true
			break
		}
	}

	if !has_role {
		return others
	}

	others = append(others, pt)
	return others
}

// Ancestors returns the ancestors of role "common" for 'pt'.
func Ancestors(pt *WOFPlacetype) []*WOFPlacetype {
	return AncestorsForRoles(pt, []string{"common"})
}

// AncestorsForRoles returns the ancestors matching any role in 'roles' for 'pt'.
func AncestorsForRoles(pt *WOFPlacetype, roles []string) []*WOFPlacetype {

	str_roles := strings.Join(roles, "-")
	key := fmt.Sprintf("%d_ancestors_%s", pt.Id, str_roles)

	v, ok := relationships.Load(key)

	if ok {
		return v.([]*WOFPlacetype)
	}

	ancestors := make([]*WOFPlacetype, 0)
	ancestors = fetchAncestors(pt, roles, ancestors)

	relationships.Store(key, ancestors)
	return ancestors
}

func fetchAncestors(pt *WOFPlacetype, roles []string, ancestors []*WOFPlacetype) []*WOFPlacetype {

	for _, id := range pt.Parent {

		parent, _ := GetPlacetypeById(id)

		role_ok := false

		for _, r := range roles {

			if r == parent.Role {
				role_ok = true
				break
			}
		}

		if !role_ok {
			continue
		}

		append_ok := true

		for _, a := range ancestors {

			if a.Id == parent.Id {
				append_ok = false
				break
			}
		}

		if append_ok {

			has_grandparent := false
			offset := -1

			for _, gpid := range parent.Parent {

				for idx, a := range ancestors {

					if a.Id == gpid {
						offset = idx
						has_grandparent = true
						break
					}
				}

				if has_grandparent {
					break
				}
			}

			// log.Printf("APPEND %s < %s GP: %t (%d)\n", parent.Name, pt.Name, has_grandparent, offset)

			if has_grandparent {

				// log.Println("WTF 1", len(ancestors))

				tail := ancestors[offset+1:]
				ancestors = ancestors[0:offset]

				ancestors = append(ancestors, parent)

				for _, a := range tail {
					ancestors = append(ancestors, a)
				}

			} else {
				ancestors = append(ancestors, parent)
			}
		}

		ancestors = fetchAncestors(parent, roles, ancestors)
	}

	return ancestors
}
