package flags

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
)

func AppendIndexingFlags(fs *flag.FlagSet) error {

	modes := emitter.Schemes()
	sort.Strings(modes)

	valid_modes := strings.Join(modes, ", ")
	desc_modes := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported schemes are: %s.", valid_modes)

	fs.String(IteratorURIFlag, "repo://", desc_modes)

	return nil
}

func ValidateIndexingFlags(fs *flag.FlagSet) error {

	_, err := lookup.StringVar(fs, IteratorURIFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", IteratorURIFlag, err)
	}

	return nil
}
