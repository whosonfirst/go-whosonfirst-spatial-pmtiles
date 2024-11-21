package flags

// THERE IS A GOOD CHANGE THIS WILL BE MOVED IN TO THE whosonfirst/go-whosonfirst-iterate PACKAGE...

import (
	"fmt"
	"sort"
	"strings"

	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
)

const SEP_FRAGMENT string = "#"
const SEP_PIPE string = "|"
const SEP_SPACE string = " "
const SEP_CSV string = ","

// IteratorURIDescription returns a string describing the use the of the `IteratorURIFlag` flag.
func IteratorURIFlagDescription() string {

	modes := emitter.Schemes()
	sort.Strings(modes)

	valid_modes := strings.Join(modes, ", ")

	desc := fmt.Sprintf("URIs take the form of {ITERATOR_URI} + \"#\" + {PIPE-SEPARATED LIST OF ITERATOR SOURCES}. Where {ITERATOR_URI} is expected to be a registered whosonfirst/go-whosonfirst-iterate/v2 iterator (emitter) URI and {ITERATOR SOURCES} are valid input paths for that iterator. Supported whosonfirst/go-whosonfirst-iterate/v2 iterator schemes are: %s.", valid_modes)

	return desc
}

// IteratorURIFlag parses a string in to components necessary for use with a `whosonfirst/go-whosonfirst-iterate/v2/iterator.Iterator` instance.
// Flags are expected to take the form of:
//
//	{ITERATOR_URI_STRING} + "#" + {PIPE_SEPARATED_LIST_OF_ITERATOR_SOURCES}
//
// Where:
// * `{ITERATOR_URI_STRING}` is a URI that can be passed to the `iterator.NewIterator` constructor.
// * `{PIPE_SEPARATED_LIST_OF_ITERATOR_SOURCES}` is a list of one or more paths (URIs) to be passed to the `iterator.IterateURIs` method.
type IteratorURIFlag struct {
	iter_uri     string
	iter_sources []string
}

// Set parses 'value' in to components necessary for use with a `whosonfirst/go-whosonfirst-iterate/v2/iterator.Iterator` instance.
// 'value' is expected to take the form of:
//
//	{ITERATOR_URI_STRING} + "#" + {PIPE_SEPARATED_LIST_OF_ITERATOR_SOURCES}
//
// Where:
// * `{ITERATOR_URI_STRING}` is a URI that can be passed to the `iterator.NewIterator` constructor.
// * `{PIPE_SEPARATED_LIST_OF_ITERATOR_SOURCES}` is a list of one or more paths (URIs) to be passed to the `iterator.IterateURIs` method.
func (fl *IteratorURIFlag) Set(value string) error {

	parts := strings.Split(value, SEP_FRAGMENT)

	if len(parts) != 2 {
		return fmt.Errorf("Invalid iterator URI")
	}

	iter_uri := parts[0]
	iter_sources := strings.Split(parts[1], SEP_PIPE)

	if len(iter_sources) == 0 {
		return fmt.Errorf("Iterator URI missing sources")
	}

	fl.iter_uri = iter_uri
	fl.iter_sources = iter_sources

	return nil
}

// Key returns the value of the `{ITERATOR_URI_STRING}` string.
func (fl *IteratorURIFlag) Key() string {
	return fl.iter_uri
}

// Values returns the list of URIs defined in the `{PIPE_SEPARATED_LIST_OF_ITERATOR_SOURCES}` string.
func (fl *IteratorURIFlag) Value() interface{} {
	return fl.iter_sources
}

// String returns value of the flag.
func (fl *IteratorURIFlag) String() string {

	str_sources := strings.Join(fl.iter_sources, SEP_PIPE)

	parts := []string{
		fl.iter_uri,
		str_sources,
	}

	return strings.Join(parts, SEP_FRAGMENT)
}

// MultiIteratorURIFlag is a flag that can hold multiple `IteratorURIFlag` instances.
type MultiIteratorURIFlag []*IteratorURIFlag

func (fl *MultiIteratorURIFlag) Set(value string) error {

	iter_fl := new(IteratorURIFlag)

	err := iter_fl.Set(value)

	if err != nil {
		return err
	}

	*fl = append(*fl, iter_fl)
	return nil
}

func (fl *MultiIteratorURIFlag) Key() string {
	return ""
}

func (fl *MultiIteratorURIFlag) Value() interface{} {
	return fl
}

func (fl *MultiIteratorURIFlag) String() string {

	str_flags := make([]string, len(*fl))

	for i, iter_fl := range *fl {
		str_flags[i] = iter_fl.String()
	}

	return strings.Join(str_flags, SEP_SPACE)
}

func (fl *MultiIteratorURIFlag) AsMap() map[string][]string {

	iter_map := make(map[string][]string)

	for _, iter_fl := range *fl {

		iter_uri := iter_fl.Key()
		iter_sources := iter_fl.Value().([]string)

		iter_map[iter_uri] = iter_sources
	}

	return iter_map
}

// MultiCSVIteratorURIFlag is a flag that can hold multiple `IteratorURIFlag` instances
// defined using a single comma-separated string.
type MultiCSVIteratorURIFlag []*IteratorURIFlag

func (fl *MultiCSVIteratorURIFlag) Set(value string) error {

	for _, str_fl := range strings.Split(value, SEP_CSV) {

		iter_fl := new(MultiIteratorURIFlag)

		err := iter_fl.Set(str_fl)

		if err != nil {
			return err
		}

		for _, iter_v_fl := range *iter_fl {
			*fl = append(*fl, iter_v_fl)
		}
	}

	return nil
}

func (fl *MultiCSVIteratorURIFlag) Key() string {
	return ""
}

func (fl *MultiCSVIteratorURIFlag) Value() interface{} {
	return fl
}

func (fl *MultiCSVIteratorURIFlag) String() string {

	str_flags := make([]string, len(*fl))

	for i, iter_fl := range *fl {
		str_flags[i] = iter_fl.String()
	}

	return strings.Join(str_flags, SEP_CSV)
}

func (fl *MultiCSVIteratorURIFlag) AsMap() map[string][]string {

	iter_map := make(map[string][]string)

	for _, iter_fl := range *fl {

		iter_uri := iter_fl.Key()
		iter_sources := iter_fl.Value().([]string)

		iter_map[iter_uri] = iter_sources
	}

	return iter_map
}
