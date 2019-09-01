# helm kelp

A Helm plugin which allows you to apply `kustomize` on helm charts before processing the templates. This allows the user to have `kustomize` to be used on helm charts so that chart overlays can be created without having to process the charts via helm first. 

## Installation

`cd $GOPATH/src/github.com`
`mkdir waveywaves && cd waveywaves`
`git clone https://github.com/waveywaves/helm-kelp && cd helm-kelp`
`make`

## Usage 

`helm kelp examples/alpine`

### TODO

* optional `kelp` configuration in the chart home
* `kelp` configuration allowing users to reuse helm template values after applying the `kustomization`.
* Use `go templates` on the helm charts for parsing instead of just `regexp`.

