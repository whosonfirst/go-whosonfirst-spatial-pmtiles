module github.com/whosonfirst/go-whosonfirst-spatial-pmtiles

go 1.23.3

require (
	github.com/aaronland/go-roster v1.0.0
	github.com/aaronland/gocloud-blob v0.4.0
	github.com/aaronland/gocloud-docstore v0.0.8
	github.com/json-iterator/go v1.1.12
	github.com/paulmach/orb v0.11.1
	github.com/protomaps/go-pmtiles v1.22.3
	github.com/sfomuseum/go-database v0.0.10
	github.com/tidwall/gjson v1.18.0
	github.com/tidwall/sjson v1.2.5
	github.com/whosonfirst/go-ioutil v1.0.2
	github.com/whosonfirst/go-reader v1.0.2
	github.com/whosonfirst/go-whosonfirst-feature v0.0.28
	github.com/whosonfirst/go-whosonfirst-spatial v0.11.1
	github.com/whosonfirst/go-whosonfirst-spatial-grpc v0.2.1
	github.com/whosonfirst/go-whosonfirst-spatial-sqlite v0.12.0
	github.com/whosonfirst/go-whosonfirst-spatial-www v0.4.0
	github.com/whosonfirst/go-whosonfirst-spr/v2 v2.3.7
	github.com/whosonfirst/go-whosonfirst-uri v1.3.0
	gocloud.dev v0.40.0
	modernc.org/sqlite v1.34.3
)

require (
	cloud.google.com/go v0.115.0 // indirect
	cloud.google.com/go/auth v0.8.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.4 // indirect
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	cloud.google.com/go/iam v1.1.13 // indirect
	cloud.google.com/go/storage v1.43.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.14.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.3.2 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/RoaringBitmap/roaring v1.5.0 // indirect
	github.com/aaronland/go-artisanal-integers v0.9.1 // indirect
	github.com/aaronland/go-aws-auth v1.7.0 // indirect
	github.com/aaronland/go-aws-dynamodb v0.3.0 // indirect
	github.com/aaronland/go-aws-session v0.2.1 // indirect
	github.com/aaronland/go-brooklynintegers-api v1.2.7 // indirect
	github.com/aaronland/go-http-bootstrap v0.5.0 // indirect
	github.com/aaronland/go-http-leaflet v0.5.0 // indirect
	github.com/aaronland/go-http-maps v0.4.0 // indirect
	github.com/aaronland/go-http-ping/v2 v2.0.0 // indirect
	github.com/aaronland/go-http-rewrite v1.1.0 // indirect
	github.com/aaronland/go-http-sanitize v0.0.8 // indirect
	github.com/aaronland/go-http-server v1.5.0 // indirect
	github.com/aaronland/go-http-static v0.0.3 // indirect
	github.com/aaronland/go-json-query v0.1.5 // indirect
	github.com/aaronland/go-pagination v0.3.0 // indirect
	github.com/aaronland/go-pagination-sql v0.2.0 // indirect
	github.com/aaronland/go-pool/v2 v2.0.0 // indirect
	github.com/aaronland/go-string v1.0.0 // indirect
	github.com/aaronland/go-uid v0.4.0 // indirect
	github.com/aaronland/go-uid-artisanal v0.0.4 // indirect
	github.com/aaronland/go-uid-proxy v0.3.0 // indirect
	github.com/aaronland/go-uid-whosonfirst v0.0.5 // indirect
	github.com/akrylysov/algnhsa v1.1.0 // indirect
	github.com/aws/aws-lambda-go v1.47.0 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/aws/aws-sdk-go-v2 v1.32.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.28.3 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.44 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.19 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.27.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.34.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.37.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.68.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.55.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.32.4 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dhconnelly/rtreego v1.2.0 // indirect
	github.com/dominikbraun/graph v0.23.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/g8rswimmer/error-chain v1.0.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/google/wire v0.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jtacoma/uritemplates v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/paulmach/go.geojson v1.4.0 // indirect
	github.com/paulmach/protoscan v0.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/schollz/progressbar/v3 v3.13.1 // indirect
	github.com/sfomuseum/go-edtf v1.2.1 // indirect
	github.com/sfomuseum/go-flags v0.10.0 // indirect
	github.com/sfomuseum/go-http-auth v0.12.0 // indirect
	github.com/sfomuseum/go-http-protomaps v0.3.0 // indirect
	github.com/sfomuseum/go-http-rollup v0.0.3 // indirect
	github.com/sfomuseum/go-sfomuseum-mapshaper v0.0.3 // indirect
	github.com/sfomuseum/go-sfomuseum-pmtiles v1.4.1 // indirect
	github.com/sfomuseum/go-template v1.10.0 // indirect
	github.com/sfomuseum/go-timings v1.4.0 // indirect
	github.com/sfomuseum/iso8601duration v1.1.0 // indirect
	github.com/sfomuseum/runtimevar v1.2.0 // indirect
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e // indirect
	github.com/tdewolff/minify/v2 v2.20.32 // indirect
	github.com/tdewolff/parse/v2 v2.7.14 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/whosonfirst/go-rfc-5646 v0.1.0 // indirect
	github.com/whosonfirst/go-sanitize v0.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-crawl v0.2.2 // indirect
	github.com/whosonfirst/go-whosonfirst-database v0.0.8 // indirect
	github.com/whosonfirst/go-whosonfirst-export/v2 v2.8.3 // indirect
	github.com/whosonfirst/go-whosonfirst-flags v0.5.2 // indirect
	github.com/whosonfirst/go-whosonfirst-format v0.4.1 // indirect
	github.com/whosonfirst/go-whosonfirst-id v1.2.5 // indirect
	github.com/whosonfirst/go-whosonfirst-iterate/v2 v2.5.0 // indirect
	github.com/whosonfirst/go-whosonfirst-names v0.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-placetypes v0.7.3 // indirect
	github.com/whosonfirst/go-whosonfirst-reader v1.0.2 // indirect
	github.com/whosonfirst/go-whosonfirst-sources v0.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-spelunker v0.0.5 // indirect
	github.com/whosonfirst/go-whosonfirst-spr-geojson v0.0.8 // indirect
	github.com/whosonfirst/go-whosonfirst-sqlite-spr/v2 v2.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-writer/v3 v3.1.4 // indirect
	github.com/whosonfirst/go-writer-featurecollection/v3 v3.0.0-20220916180959-42588e308a3e // indirect
	github.com/whosonfirst/go-writer/v3 v3.1.1 // indirect
	github.com/whosonfirst/walk v0.0.2 // indirect
	go.mongodb.org/mongo-driver v1.11.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.uber.org/ratelimit v0.3.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/term v0.24.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	golang.org/x/xerrors v0.0.0-20240716161551-93cc26a95ae9 // indirect
	google.golang.org/api v0.191.0 // indirect
	google.golang.org/genproto v0.0.0-20240812133136-8ffd90a71988 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	modernc.org/gc/v3 v3.0.0-20240107210532-573471604cb6 // indirect
	modernc.org/libc v1.55.3 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
	zombiezen.com/go/sqlite v1.1.2 // indirect
)
