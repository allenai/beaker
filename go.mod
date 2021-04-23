module github.com/allenai/beaker

go 1.13

require (
	github.com/allenai/bytefmt v0.1.1
	github.com/beaker/client v0.0.0-20210421233118-5dcaf8303b8d
	github.com/beaker/fileheap v0.0.0-20210213001550-3d3932012952
	github.com/beaker/runtime v0.0.0-20210423161731-bafd51af5b2e
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.3+incompatible
	github.com/fatih/color v1.10.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	golang.org/x/net v0.0.0-20210331060903-cb1fcc7394e5 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/spf13/viper => ./viperstub
