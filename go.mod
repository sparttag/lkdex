module github.com/lianxiangcloud/lkdex

go 1.12

replace (
	github.com/NebulousLabs/go-upnp => github.com/lianxiangcloud/go-upnp v0.0.0-20190905032046-65768e0b268c
	github.com/go-interpreter/wagon => github.com/xunleichain/wagon v0.5.3
	gopkg.in/sourcemap.v1 => github.com/go-sourcemap/sourcemap v1.0.5
)

require (
	github.com/bouk/monkey v1.0.1
	github.com/golang/mock v1.3.1
	github.com/jinzhu/gorm v1.9.11
	github.com/lianxiangcloud/linkchain v0.1.2
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	github.com/xunleichain/tc-wasm v0.3.5
	gopkg.in/h2non/gock.v1 v1.0.15
)
