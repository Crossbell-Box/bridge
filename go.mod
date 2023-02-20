module github.com/Crossbell-Box/bridge

replace github.com/axieinfinity/bridge-v2 => ./

replace github.com/axieinfinity/bridge-core => github.com/Crossbell-Box/bridge-core v0.0.0-20230218061833-f3e2a37474a0

replace github.com/axieinfinity/bridge-migrations => github.com/Crossbell-Box/bridge-migrations v0.0.0-20230211122754-a6378691b5cf

replace github.com/btcsuite/btcd => github.com/btcsuite/btcd/chaincfg/chainhash v1.0.2

go 1.17

require (
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5
	github.com/hashicorp/go-bexpr v0.1.10
	github.com/mattn/go-colorable v0.1.11
	github.com/mattn/go-isatty v0.0.14
	golang.org/x/crypto v0.1.0 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
	gorm.io/gorm v1.24.1
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/ashwanthkumar/slack-go-webhook v0.0.0-20200209025033-430dd4e66960
	github.com/axieinfinity/bridge-contracts v0.0.0-20230111072442-13fb0177332d
	github.com/axieinfinity/bridge-core v0.1.2-0.20221221074635-375d6a0ea127
	github.com/axieinfinity/bridge-migrations v0.0.0-00010101000000-000000000000
	github.com/axieinfinity/bridge-v2 v0.0.0-00010101000000-000000000000
	github.com/axieinfinity/ronin-kms-client v0.0.0-20220805072849-960e04981b70
	github.com/ethereum/go-ethereum v1.10.26
	github.com/miguelmota/go-solidity-sha3 v0.1.1
	gorm.io/driver/postgres v1.4.5
)

require (
	// github.com/Crossbell-Box/bridge-contracts v0.0.0-20230116152603-ae6c94b86c68 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/go-gormigrate/gormigrate/v2 v2.0.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.2.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/pointerstructure v1.2.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/rivo/uniseg v0.4.2 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.5.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20221107162902-2d387536bcdd // indirect
	google.golang.org/grpc v1.50.1 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gorm.io/driver/sqlite v1.4.3 // indirect
)
