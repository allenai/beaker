module github.com/allenai/beaker

go 1.13

require (
	github.com/allenai/bytefmt v0.1.0
	github.com/beaker/client v0.0.0-20210325140910-d37f84b8e023
	github.com/beaker/fileheap v0.0.0-20210213001550-3d3932012952
	github.com/beaker/runtime v0.0.0-20210330202441-23e0e6a75442
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.3+incompatible
	github.com/fatih/color v1.10.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	golang.org/x/net v0.0.0-20210330210036-cd0ac97f97f6 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/spf13/viper => ./viperstub
