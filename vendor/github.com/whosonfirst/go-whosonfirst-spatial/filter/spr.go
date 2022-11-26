package filter

import (
	"fmt"
	"github.com/whosonfirst/go-sanitize"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/date"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
	"github.com/whosonfirst/go-whosonfirst-flags/geometry"
	"github.com/whosonfirst/go-whosonfirst-flags/placetypes"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	_ "log"
	"strconv"
	"strings"
)

var sanitizeOpts *sanitize.Options

func init() {
	sanitizeOpts = sanitize.DefaultOptions()
}

type SPRInputs struct {
	Placetypes          []string
	IsCurrent           []int64
	IsCeased            []int64
	IsDeprecated        []int64
	IsSuperseded        []int64
	IsSuperseding       []int64
	Geometries          []string
	AlternateGeometries []string
	InceptionDate       string
	CessationDate       string
}

type SPRFilter struct {
	spatial.Filter
	Placetypes          []flags.PlacetypeFlag
	Current             []flags.ExistentialFlag
	Deprecated          []flags.ExistentialFlag
	Ceased              []flags.ExistentialFlag
	Superseded          []flags.ExistentialFlag
	Superseding         []flags.ExistentialFlag
	AlternateGeometry   flags.AlternateGeometryFlag
	AlternateGeometries []flags.AlternateGeometryFlag
	InceptionDate       flags.DateFlag
	CessationDate       flags.DateFlag
}

func (f *SPRFilter) MatchesInception(fl flags.DateFlag) bool {

	return f.InceptionDate.MatchesAny(fl)
}

func (f *SPRFilter) MatchesCessation(fl flags.DateFlag) bool {

	return f.CessationDate.MatchesAny(fl)
}

func (f *SPRFilter) HasPlacetypes(fl flags.PlacetypeFlag) bool {

	for _, p := range f.Placetypes {

		if p.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsCurrent(fl flags.ExistentialFlag) bool {

	for _, e := range f.Current {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsDeprecated(fl flags.ExistentialFlag) bool {

	for _, e := range f.Deprecated {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsCeased(fl flags.ExistentialFlag) bool {

	for _, e := range f.Ceased {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsSuperseded(fl flags.ExistentialFlag) bool {

	for _, e := range f.Superseded {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsSuperseding(fl flags.ExistentialFlag) bool {

	for _, e := range f.Superseding {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsAlternateGeometry(fl flags.AlternateGeometryFlag) bool {

	return f.AlternateGeometry.MatchesAny(fl)
}

func (f *SPRFilter) HasAlternateGeometry(fl flags.AlternateGeometryFlag) bool {

	for _, a := range f.AlternateGeometries {

		if a.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func NewSPRInputs() (*SPRInputs, error) {

	i := SPRInputs{
		Placetypes:          make([]string, 0),
		IsCurrent:           make([]int64, 0),
		IsDeprecated:        make([]int64, 0),
		IsCeased:            make([]int64, 0),
		IsSuperseded:        make([]int64, 0),
		IsSuperseding:       make([]int64, 0),
		Geometries:          make([]string, 0),
		AlternateGeometries: make([]string, 0),
		InceptionDate:       "",
		CessationDate:       "",
	}

	return &i, nil
}

func NewSPRFilter() (*SPRFilter, error) {

	null_pt, _ := placetypes.NewNullFlag()
	null_ex, _ := existential.NewNullFlag()
	null_alt, _ := geometry.NewNullAlternateGeometryFlag()
	null_dt, _ := date.NewNullDateFlag()

	col_pt := []flags.PlacetypeFlag{null_pt}
	col_ex := []flags.ExistentialFlag{null_ex}
	col_alt := []flags.AlternateGeometryFlag{null_alt}

	f := SPRFilter{
		Placetypes:          col_pt,
		Current:             col_ex,
		Deprecated:          col_ex,
		Ceased:              col_ex,
		Superseded:          col_ex,
		Superseding:         col_ex,
		AlternateGeometry:   null_alt,
		AlternateGeometries: col_alt,
		InceptionDate:       null_dt,
		CessationDate:       null_dt,
	}

	return &f, nil
}

func NewSPRFilterFromInputs(inputs *SPRInputs) (spatial.Filter, error) {

	f, err := NewSPRFilter()

	if err != nil {
		return nil, err
	}

	if len(inputs.Placetypes) != 0 {

		possible, err := placetypeFlags(inputs.Placetypes)

		if err != nil {
			return nil, err
		}

		f.Placetypes = possible
	}

	if inputs.InceptionDate != "" {

		fl, err := date.NewEDTFDateFlag(inputs.InceptionDate)

		if err != nil {
			return nil, err
		}

		f.InceptionDate = fl
	}

	if inputs.CessationDate != "" {

		fl, err := date.NewEDTFDateFlag(inputs.CessationDate)

		if err != nil {
			return nil, err
		}

		f.CessationDate = fl
	}

	if len(inputs.IsCurrent) != 0 {

		possible, err := existentialFlags(inputs.IsCurrent)

		if err != nil {
			return nil, err
		}

		f.Current = possible
	}

	if len(inputs.IsDeprecated) != 0 {

		possible, err := existentialFlags(inputs.IsDeprecated)

		if err != nil {
			return nil, err
		}

		f.Deprecated = possible
	}

	if len(inputs.IsCeased) != 0 {

		possible, err := existentialFlags(inputs.IsCeased)

		if err != nil {
			return nil, err
		}

		f.Ceased = possible
	}

	if len(inputs.IsSuperseded) != 0 {

		possible, err := existentialFlags(inputs.IsSuperseded)

		if err != nil {
			return nil, err
		}

		f.Superseded = possible
	}

	if len(inputs.IsSuperseding) != 0 {

		possible, err := existentialFlags(inputs.IsSuperseding)

		if err != nil {
			return nil, err
		}

		f.Superseding = possible
	}

	if len(inputs.Geometries) != 0 {

		geoms := inputs.Geometries[0]

		switch geoms {
		case "all":
			// pass
		case "alt", "alternate":

			af, err := geometry.NewIsAlternateGeometryFlag(true)

			if err != nil {
				return nil, fmt.Errorf("Failed to create alternate geometry flag, %v", err)
			}

			f.AlternateGeometry = af

		case "default":

			af, err := geometry.NewIsAlternateGeometryFlag(false)

			if err != nil {
				return nil, fmt.Errorf("Failed to create alternate geometry flag, %v", err)
			}

			f.AlternateGeometry = af

		default:
			fmt.Errorf("Invalid geometries flag")
		}

	}

	if len(inputs.AlternateGeometries) != 0 {

		possible, err := hasAlternateGeometryFlags(inputs.AlternateGeometries)

		if err != nil {
			return nil, err
		}

		f.AlternateGeometries = possible
	}

	return f, nil
}

func dateFlags(inputs []string) ([]flags.DateFlag, error) {

	possible := make([]flags.DateFlag, 0)

	for _, raw := range inputs {

		candidates, err := stringList(raw, ",")

		if err != nil {
			return nil, err
		}

		for _, edtf_str := range candidates {

			fl, err := date.NewEDTFDateFlag(edtf_str)

			if err != nil {
				return nil, err
			}

			possible = append(possible, fl)
		}
	}

	return possible, nil
}

func placetypeFlags(inputs []string) ([]flags.PlacetypeFlag, error) {

	possible := make([]flags.PlacetypeFlag, 0)

	for _, raw := range inputs {

		candidates, err := stringList(raw, ",")

		if err != nil {
			return nil, err
		}

		for _, pt := range candidates {

			fl, err := placetypes.NewPlacetypeFlag(pt)

			if err != nil {
				return nil, err
			}

			possible = append(possible, fl)
		}
	}

	return possible, nil
}

func existentialFlags(inputs []int64) ([]flags.ExistentialFlag, error) {

	possible := make([]flags.ExistentialFlag, 0)

	for _, i := range inputs {

		fl, err := existential.NewKnownUnknownFlag(i)

		if err != nil {
			return nil, err
		}

		possible = append(possible, fl)
	}

	return possible, nil
}

func hasAlternateGeometryFlags(input []string) ([]flags.AlternateGeometryFlag, error) {

	possible := make([]flags.AlternateGeometryFlag, 0)

	for _, raw := range input {

		candidates, err := stringList(raw, ",")

		if err != nil {
			return nil, err
		}

		for _, alt_label := range candidates {

			uri_str := geometry.DummyAlternateGeometryURIWithLabel(alt_label)

			fl, err := geometry.NewAlternateGeometryFlag(uri_str)

			if err != nil {
				return nil, err
			}

			possible = append(possible, fl)
		}
	}

	return possible, nil
}

func stringList(raw string, sep string) ([]string, error) {

	str, err := sanitize.SanitizeString(raw, sanitizeOpts)

	if err != nil {
		return nil, err
	}

	str_list := make([]string, 0)

	str = strings.Trim(str, " ")

	for _, str_i := range strings.Split(str, sep) {

		str_i = strings.Trim(str_i, " ")

		if str_i == "" {
			continue
		}

		str_list = append(str_list, str_i)
	}

	return str_list, nil
}

func int64List(raw string, sep string) ([]int64, error) {

	str, err := sanitize.SanitizeString(raw, sanitizeOpts)

	if err != nil {
		return nil, err
	}

	int64_list := make([]int64, 0)

	str = strings.Trim(str, " ")

	for _, str_i := range strings.Split(str, sep) {

		str_i = strings.Trim(str_i, " ")

		if str_i == "" {
			continue
		}

		i, err := strconv.ParseInt(str_i, 10, 64)

		if err != nil {
			return nil, err
		}

		int64_list = append(int64_list, i)
	}

	return int64_list, nil
}
