module github.com/whosonfirst/go-whosonfirst-spatial-pmtiles

go 1.24.5

// because the default protomaps/go-pmtiles go.mod triggers import errors involving azure bindings

replace github.com/protomaps/go-pmtiles v1.28.0 => github.com/sfomuseum/go-pmtiles v0.0.0-20250714215437-ff4cb74cab97

require (
	github.com/aaronland/go-roster v1.0.0
	github.com/aaronland/gocloud-blob v0.6.2
	github.com/aaronland/gocloud-docstore v0.0.9
	github.com/json-iterator/go v1.1.12
	github.com/paulmach/orb v0.11.1
	github.com/protomaps/go-pmtiles v1.28.0
	github.com/sfomuseum/go-database v0.0.14
	github.com/tidwall/gjson v1.18.0
	github.com/tidwall/sjson v1.2.5
	github.com/whosonfirst/go-ioutil v1.0.2
	github.com/whosonfirst/go-reader/v2 v2.0.0
	github.com/whosonfirst/go-whosonfirst-feature v0.0.29
	github.com/whosonfirst/go-whosonfirst-spatial v0.18.1
	github.com/whosonfirst/go-whosonfirst-spatial-grpc v0.3.0
	github.com/whosonfirst/go-whosonfirst-spatial-sqlite v0.15.2
	github.com/whosonfirst/go-whosonfirst-spatial-www v0.7.2
	github.com/whosonfirst/go-whosonfirst-spr/v2 v2.3.7
	github.com/whosonfirst/go-whosonfirst-uri v1.3.0
	gocloud.dev v0.43.0
	modernc.org/sqlite v1.38.2
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.4 // indirect
	cloud.google.com/go/auth v0.16.3 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	cloud.google.com/go/storage v1.55.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.1 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.29.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.53.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.53.0 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/RoaringBitmap/roaring v1.9.4 // indirect
	github.com/aaronland/go-artisanal-integers v0.9.1 // indirect
	github.com/aaronland/go-aws-auth v1.7.0 // indirect
	github.com/aaronland/go-aws-auth/v2 v2.0.1 // indirect
	github.com/aaronland/go-aws-dynamodb v0.4.2 // indirect
	github.com/aaronland/go-aws-session v0.2.1 // indirect
	github.com/aaronland/go-brooklynintegers-api v1.2.10 // indirect
	github.com/aaronland/go-http-maps/v2 v2.0.0 // indirect
	github.com/aaronland/go-http-ping/v2 v2.0.0 // indirect
	github.com/aaronland/go-http-sanitize v0.0.8 // indirect
	github.com/aaronland/go-http-server/v2 v2.0.1 // indirect
	github.com/aaronland/go-json-query v0.1.6 // indirect
	github.com/aaronland/go-pagination v0.3.0 // indirect
	github.com/aaronland/go-pagination-sql v0.2.0 // indirect
	github.com/aaronland/go-pool/v2 v2.0.0 // indirect
	github.com/aaronland/go-string v1.0.0 // indirect
	github.com/aaronland/go-uid v0.5.0 // indirect
	github.com/aaronland/go-uid-artisanal v0.0.5 // indirect
	github.com/aaronland/go-uid-proxy v0.4.1 // indirect
	github.com/aaronland/go-uid-whosonfirst v0.0.7 // indirect
	github.com/akrylysov/algnhsa v1.1.0 // indirect
	github.com/aws/aws-lambda-go v1.49.0 // indirect
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.6 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.11 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.18 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.71 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.33 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.84 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.29.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.44.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.43.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.84.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.60.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.34.1 // indirect
	github.com/aws/smithy-go v1.22.4 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/dhconnelly/rtreego v1.2.0 // indirect
	github.com/dominikbraun/graph v0.23.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/g8rswimmer/error-chain v1.0.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/google/wire v0.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jtacoma/uritemplates v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.30 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/paulmach/protoscan v0.2.1 // indirect
	github.com/peterstace/simplefeatures v0.54.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/schollz/progressbar/v3 v3.18.0 // indirect
	github.com/sfomuseum/go-edtf v1.2.1 // indirect
	github.com/sfomuseum/go-flags v0.11.0 // indirect
	github.com/sfomuseum/go-http-auth v1.2.0 // indirect
	github.com/sfomuseum/go-sfomuseum-mapshaper v0.0.4 // indirect
	github.com/sfomuseum/go-timings v1.4.0 // indirect
	github.com/sfomuseum/iso8601duration v1.1.0 // indirect
	github.com/sfomuseum/runtimevar v1.3.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/tidwall/geoindex v1.4.4 // indirect
	github.com/tidwall/geojson v1.4.5 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/rtree v1.3.1 // indirect
	github.com/whosonfirst/go-reader v1.1.0 // indirect
	github.com/whosonfirst/go-rfc-5646 v0.1.0 // indirect
	github.com/whosonfirst/go-sanitize v0.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-database v0.0.9 // indirect
	github.com/whosonfirst/go-whosonfirst-export/v3 v3.0.4 // indirect
	github.com/whosonfirst/go-whosonfirst-flags v0.5.2 // indirect
	github.com/whosonfirst/go-whosonfirst-format v1.0.1 // indirect
	github.com/whosonfirst/go-whosonfirst-id v1.3.1 // indirect
	github.com/whosonfirst/go-whosonfirst-iterate/v3 v3.2.0 // indirect
	github.com/whosonfirst/go-whosonfirst-names v0.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-placetypes v0.8.0 // indirect
	github.com/whosonfirst/go-whosonfirst-reader/v2 v2.0.0 // indirect
	github.com/whosonfirst/go-whosonfirst-sources v0.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-spelunker v0.0.6 // indirect
	github.com/whosonfirst/go-whosonfirst-spr-geojson/v2 v2.0.0 // indirect
	github.com/whosonfirst/go-whosonfirst-sqlite-spr/v2 v2.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-validate v0.6.1 // indirect
	github.com/whosonfirst/go-whosonfirst-writer/v3 v3.1.7 // indirect
	github.com/whosonfirst/go-writer-featurecollection/v3 v3.0.2 // indirect
	github.com/whosonfirst/go-writer/v3 v3.1.1 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	go.mongodb.org/mongo-driver v1.11.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.37.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.62.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.uber.org/ratelimit v0.3.1 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.242.0 // indirect
	google.golang.org/genproto v0.0.0-20250715232539-7130f93afb79 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250715232539-7130f93afb79 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250715232539-7130f93afb79 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	modernc.org/libc v1.66.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	zombiezen.com/go/sqlite v1.4.2 // indirect
)
