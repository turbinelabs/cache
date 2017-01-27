
[//]: # ( Copyright 2017 Turbine Labs, Inc.                                   )
[//]: # ( you may not use this file except in compliance with the License.    )
[//]: # ( You may obtain a copy of the License at                             )
[//]: # (                                                                     )
[//]: # (     http://www.apache.org/licenses/LICENSE-2.0                      )
[//]: # (                                                                     )
[//]: # ( Unless required by applicable law or agreed to in writing, software )
[//]: # ( distributed under the License is distributed on an "AS IS" BASIS,   )
[//]: # ( WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or     )
[//]: # ( implied. See the License for the specific language governing        )
[//]: # ( permissions and limitations under the License.                      )

# turbinelabs/cache

[![GoDoc](https://https://godoc.org/github.com/turbinelabs/cache?status.svg)](https://https://godoc.org/github.com/turbinelabs/cache)
[https://circleci.com/gh/turbinelabs/cache](![CircleCI`](https://circleci.com/gh/turbinelabs/cache.svg?style=svg))

The cache project provides a simple Cache interface, with several concrete
implementations.

## Requirements

- Go 1.7.4 or later (previous versions may work, but we don't build or test against them)

## Dependencies

The cache project depends on our [nonstdlib package](https://github.com/turbinelabs/nonstdlib);
the tests depend on our [test package](https://github.com/turbinelabs/test).
It should always be safe to use HEAD of all master branches of Turbine Labs
open source projects together, or to vendor them with the same git tag.

A [gomock](https://github.com/golang/mock)-based MockCache is provided.

Additionally, we vendor
[github.com/hashicorp/golang-lru](https://github.com/hashicorp/golang-lru).
This should be considered an opaque implementation detail,
see [Vendoring](http://github.com/turbinelabs/developer/blob/master/README.md#vendoring)
for more detail.

## Install

```
go get -u github.com/turbinelabs/cache/...
```

## Clone/Test

```
mkdir -p $GOPATH/src/turbinelabs
git clone https://github.com/turbinelabs/cache.git > $GOPATH/src/turbinelabs/cache
go test github.com/turbinelabs/cache/...
```

## Godoc

[`cache`](https://godoc.org/github.com/turbinelabs/cache)

## Versioning

Please see [Versioning of Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#versioning).

## Pull Requests

Patches accepted! Please see [Contributing to Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#contributing).
