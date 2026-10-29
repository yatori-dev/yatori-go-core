// Package driver wires every platform adapter into the SDK registry.
//
// The package is imported for its side effects by the public yatori
// package. Adapters register themselves from init functions, which keeps
// the public surface free of constructor boilerplate while still allowing
// the linker to include exactly one implementation tree.
package driver

import (
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/cela"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/cqie"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/enaea"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/gongxue"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/haiqikeji"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/icve"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/ketangx"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/mooc"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/qsxt"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/ttcdw"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/weiban"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/welearn"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/xuexitong"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/yinghua"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver/zhihuishu"
)
