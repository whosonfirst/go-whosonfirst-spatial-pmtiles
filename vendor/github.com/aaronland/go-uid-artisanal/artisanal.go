package artisanal

import (
	"context"
	"fmt"
	"github.com/aaronland/go-artisanal-integers/client"
	"github.com/aaronland/go-uid"
	_ "net/url"
	"strings"
)

const ARTISANAL_SCHEME string = "artisanal"

func init() {
	ctx := context.Background()

	for _, s := range client.Schemes() {
		s = strings.Replace(s, "://", "", 1)
		uid.RegisterProvider(ctx, s, NewArtisanalProvider)
	}
}

type ArtisanalProvider struct {
	uid.Provider
	client client.Client
}

type ArtisanalUID struct {
	uid.UID
	id int64
}

func NewArtisanalProvider(ctx context.Context, uri string) (uid.Provider, error) {

	cl_uri := fmt.Sprintf(uri)

	cl, err := client.NewClient(ctx, cl_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create artisanal integer client, %w", err)
	}

	pr := &ArtisanalProvider{
		client: cl,
	}

	return pr, nil
}

func (pr *ArtisanalProvider) UID(ctx context.Context, args ...interface{}) (uid.UID, error) {
	return NewArtisanalUID(ctx, pr.client)
}

func NewArtisanalUID(ctx context.Context, args ...interface{}) (uid.UID, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("Invalid arguments")
	}

	cl, ok := args[0].(client.Client)

	if !ok {
		return nil, fmt.Errorf("Invalid client")
	}

	i, err := cl.NextInt(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new integerm %w", err)
	}

	u := &ArtisanalUID{
		id: i,
	}

	return u, nil
}

func (u *ArtisanalUID) Value() any {
	return u.id
}

func (u *ArtisanalUID) String() string {
	return fmt.Sprintf("%v", u.Value())
}
