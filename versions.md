---
layout: default
title: Versions of Falcore
---

## Versioning Falcore

Falcore uses [Semantic Versioning](http://semver.org/) 2.0.0.  You can find all [releases on github](https://github.com/fitstar/falcore/releases) and you can find the version history here.

We try to keep the master branch stable, however new APIs tend to change a lot until they stabilize for a release.

## Version History

### 1.0.2

* fixed minor bug in etag filter regarding chunked responses
* deprecated SplitHostPort.  use the version in net instead.  will be removed in v2.x.x
* deprecated TimeDiff.  use the tools in time package instead.  will be removed in v2.x.x
* improved hot restart example

### 1.0.1

* Documentation improvements.  No code changes.

### 1.0.0

* First versioned release
* Targets go1.1
