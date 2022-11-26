package geometry

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const DUMMY_ID int64 = 0

const DUMMY_PREFIX string = "dummy"

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

type AlternateGeometryFlag struct {
	flags.AlternateGeometryFlag
	is_alt bool
	label  string
}

func DummyURI() string {
	return fmt.Sprintf("%d.geojson", DUMMY_ID)
}

func DummyAlternateGeometryURI() string {
	alt_label := DummyAlternateURILabel()
	return DummyAlternateGeometryURIWithLabel(alt_label)
}

func DummyAlternateGeometryURIWithLabel(label string) string {
	return fmt.Sprintf("%d-alt-%s.geojson", DUMMY_ID, label)
}

func DummyAlternateURILabel() string {
	rand := stringWithCharset(12, charset)
	return fmt.Sprintf("%s-%s", DUMMY_PREFIX, rand)
}

// https://www.calhoun.io/creating-random-strings-in-go/

func stringWithCharset(length int, charset string) string {

	b := make([]byte, length)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

func NewIsAlternateGeometryFlagWithString(bool_str string) (flags.AlternateGeometryFlag, error) {

	is_alt, err := strconv.ParseBool(bool_str)

	if err != nil {
		return nil, err
	}

	return NewIsAlternateGeometryFlag(is_alt)
}

func NewIsAlternateGeometryFlag(is_alt bool) (flags.AlternateGeometryFlag, error) {

	uri_str := DummyURI()

	if is_alt {
		uri_str = DummyAlternateGeometryURI()
	}

	return NewAlternateGeometryFlag(uri_str)
}

func NewAlternateGeometryFlagWithLabel(label string) (flags.AlternateGeometryFlag, error) {

	uri_str := DummyAlternateGeometryURIWithLabel(label)
	return NewAlternateGeometryFlag(uri_str)
}

func NewAlternateGeometryFlagsWithLabelArray(labels ...string) ([]flags.AlternateGeometryFlag, error) {

	uris := make([]string, len(labels))

	for i, label := range labels {
		uris[i] = DummyAlternateGeometryURIWithLabel(label)
	}

	return NewAlternateGeometryFlagsArray(uris...)
}

func NewAlternateGeometryFlagsArray(uris ...string) ([]flags.AlternateGeometryFlag, error) {

	alt_flags := make([]flags.AlternateGeometryFlag, 0)

	for _, uri_str := range uris {

		fl, err := NewAlternateGeometryFlag(uri_str)

		if err != nil {
			return nil, err
		}

		alt_flags = append(alt_flags, fl)
	}

	return alt_flags, nil
}

func NewAlternateGeometryFlag(uri_str string) (flags.AlternateGeometryFlag, error) {

	_, uri_args, err := uri.ParseURI(uri_str)

	if err != nil {
		return nil, err
	}

	is_alt := uri_args.IsAlternate
	alt_label := ""

	if is_alt {

		label, err := uri_args.AltGeom.String()

		if err != nil {
			return nil, err
		}

		alt_label = label
	}

	// check label against go-whosonfirst-sources here?

	f := AlternateGeometryFlag{
		is_alt: is_alt,
		label:  alt_label,
	}

	return &f, nil
}

func (f *AlternateGeometryFlag) MatchesAny(others ...flags.AlternateGeometryFlag) bool {

	for _, o := range others {

		if f.isEqual(o) {
			return true
		}

	}

	return false
}

func (f *AlternateGeometryFlag) MatchesAll(others ...flags.AlternateGeometryFlag) bool {

	matches := 0

	for _, o := range others {

		if f.isEqual(o) {
			matches += 1
		}

	}

	if matches == len(others) {
		return true
	}

	return false
}

func (f *AlternateGeometryFlag) IsAlternateGeometry() bool {
	return f.is_alt
}

func (f *AlternateGeometryFlag) Label() string {
	return f.label
}

func (f *AlternateGeometryFlag) String() string {
	return f.Label()
}

func (f *AlternateGeometryFlag) isEqual(other flags.AlternateGeometryFlag) bool {

	if f.IsAlternateGeometry() != other.IsAlternateGeometry() {
		return false
	}

	if !strings.HasPrefix(f.Label(), DUMMY_PREFIX) {

		if f.Label() != other.Label() {
			return false
		}
	}

	return true
}
