# github.com/wptide/pkg

[![Build Status](https://travis-ci.org/wptide/pkg.svg?branch=master)](https://travis-ci.org/wptide/pkg) [![Coverage Status](https://coveralls.io/repos/github/wptide/pkg/badge.svg?branch=master)](https://coveralls.io/github/wptide/pkg?branch=master) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

This repository contains common packages that are used for the development of
Tide related services.

The packages can be added to your project using a Go package dependency manager (e.g. [glide](https://github.com/Masterminds/glide)
or [golang/dep](https://github.com/golang/dep)) or added to your GOPATH using `go get`.

```
go get github.com/wptide/pkg
```

All packages have been tested and provide 100% coverage. We expect nothing less from any contributions to this project.

Test can be run inside each package using the `go test` command or the convenience instruction in the provided `Makefile` (`make` required) by typing:
```
make test
```

This is a library of packages and is meant to be imported into Go projects.

## Props

[@rheinardkorf](https://github.com/rheinardkorf), [@valendesigns](https://github.com/valendesigns)