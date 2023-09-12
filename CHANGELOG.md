# Changelog

## [1.3.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.2.0...v1.3.0) (2023-09-12)


### Features

* add schemas endpoints ([c9be3dc](https://github.com/canonical/identity-platform-admin-ui/commit/c9be3dcf364b62c9900233f56cc448d51a5f3cf7))
* add schemas service layer and interfaces ([83917cf](https://github.com/canonical/identity-platform-admin-ui/commit/83917cf291f031a01d8882364ac3e50aedfd99e5))
* added ca-certificates package to stage-packages ([16f6683](https://github.com/canonical/identity-platform-admin-ui/commit/16f6683218d02e3c60923007a59cf36c2bb5f5d2))
* wire up schemas pkg ([513ce61](https://github.com/canonical/identity-platform-admin-ui/commit/513ce612809910c78bfc4f4647246ce4f90a1c42))


### Bug Fixes

* **deps:** update dependency @canonical/react-components to v0.47.0 ([#94](https://github.com/canonical/identity-platform-admin-ui/issues/94)) ([a2c7e03](https://github.com/canonical/identity-platform-admin-ui/commit/a2c7e0318bbeb20d3cb7aaf86122b8dd1ada49fc))
* **deps:** update dependency vanilla-framework to v4 ([#95](https://github.com/canonical/identity-platform-admin-ui/issues/95)) ([35c21ae](https://github.com/canonical/identity-platform-admin-ui/commit/35c21aea82d6c6b4cb1501d771bd249d44d476e4))
* **deps:** update dependency vanilla-framework to v4.3.0 ([#99](https://github.com/canonical/identity-platform-admin-ui/issues/99)) ([049629c](https://github.com/canonical/identity-platform-admin-ui/commit/049629c98e36c433e144595e6b4ad25f1f6872d9))
* **deps:** update go deps (minor) ([#75](https://github.com/canonical/identity-platform-admin-ui/issues/75)) ([54f9421](https://github.com/canonical/identity-platform-admin-ui/commit/54f9421d686543e552e04d9790843db83d90c103))
* **deps:** update go deps to v1.17.0 (minor) ([#71](https://github.com/canonical/identity-platform-admin-ui/issues/71)) ([472dc50](https://github.com/canonical/identity-platform-admin-ui/commit/472dc5067964d4fc183c677d82cf3a791d646cd0))
* **deps:** update go deps to v1.18.0 (minor) ([#100](https://github.com/canonical/identity-platform-admin-ui/issues/100)) ([129c7ee](https://github.com/canonical/identity-platform-admin-ui/commit/129c7eeedb143af2422063f20ae83b066543fcba))
* **deps:** update module github.com/google/uuid to v1.3.1 ([#53](https://github.com/canonical/identity-platform-admin-ui/issues/53)) ([840b068](https://github.com/canonical/identity-platform-admin-ui/commit/840b0689e8b75c3bd001e8804a1f2bec471ec47d))
* **deps:** update module go.opentelemetry.io/otel/exporters/stdout/stdouttrace to v1.17.0 ([#72](https://github.com/canonical/identity-platform-admin-ui/issues/72)) ([9fd027b](https://github.com/canonical/identity-platform-admin-ui/commit/9fd027b2a1818676b973882e67009bf494a01cd6))
* fix renovate config ([700cc51](https://github.com/canonical/identity-platform-admin-ui/commit/700cc515c2af7a56d5e5781ecd38de0cb29aaaa4))

## [1.2.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.1.0...v1.2.0) (2023-08-10)


### Features

* add idp handlers ([405bad3](https://github.com/canonical/identity-platform-admin-ui/commit/405bad314cb3b3a79b0455b74b7a123cb09818b7))
* add idp service ([4f04546](https://github.com/canonical/identity-platform-admin-ui/commit/4f04546e2a1f75f16ce36a1bea051ce012d8e44c))
* wire up main and router with new dependencies ([7c218d3](https://github.com/canonical/identity-platform-admin-ui/commit/7c218d3ea8fd9413e808afa7f54a265a3e1dec6d))


### Bug Fixes

* add otel tracing to hydra client ([64871cd](https://github.com/canonical/identity-platform-admin-ui/commit/64871cdb232a92ebb11b4ed0d05282898cdc9f9d))
* create k8s coreV1 package ([ff260f9](https://github.com/canonical/identity-platform-admin-ui/commit/ff260f927d1930fb587ac515962fe4605b2d9223))
* drop unused const ([bb3bd28](https://github.com/canonical/identity-platform-admin-ui/commit/bb3bd28a0f1df6904d5f6355b9bcc198276d8db7))
* use io pkg instead of ioutil ([909459c](https://github.com/canonical/identity-platform-admin-ui/commit/909459c1041391d6906e20ecbe9c129523c8774f))
* use new instead of & syntax ([9908ddc](https://github.com/canonical/identity-platform-admin-ui/commit/9908ddc30301816b623d0bf8e064cae1c1dd91f6))

## [1.1.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.0.0...v1.1.0) (2023-07-27)


### Features

* add hydra service ([17a3c86](https://github.com/canonical/identity-platform-admin-ui/commit/17a3c866cffcf5ef8c5f54881482ccfe2f4b4d1d))
* add identities service layer ([d619daf](https://github.com/canonical/identity-platform-admin-ui/commit/d619dafe04f3452402f488a4f75739cfdc68b2d5))
* create apis for identities kratos REST endpoints ([6da5dae](https://github.com/canonical/identity-platform-admin-ui/commit/6da5dae6f73602c80057ed20b2de7bdb06288fcb))
* create kratos client ([d009507](https://github.com/canonical/identity-platform-admin-ui/commit/d009507359360bbd1fa05b494e5db25d68721d77))


### Bug Fixes

* add jaeger propagator as ory components support only these spans for now ([5a90f83](https://github.com/canonical/identity-platform-admin-ui/commit/5a90f838f224add360c81aeaf88a66e2811a7185))
* fail if HYDRA_ADMIN_URL is not provided ([c9e1844](https://github.com/canonical/identity-platform-admin-ui/commit/c9e18449a2cef297ed34414ec1a5b88177ce9b38))
* IAM-339 - add generic response pkg ([b98a505](https://github.com/canonical/identity-platform-admin-ui/commit/b98a505ac3ababddb27a0b903842db4f73a65e1d))
* introduce otelHTTP and otelGRPC exporter for tempo ([9156892](https://github.com/canonical/identity-platform-admin-ui/commit/91568926bc441372c4b342a5cdd42433b6fbd3fb))
* only print hydra debug logs on debug ([15dc2b4](https://github.com/canonical/identity-platform-admin-ui/commit/15dc2b4ba473384569b13fcbc84ecb29cfb021d4))
* wire up new kratos endpoints ([1d881a7](https://github.com/canonical/identity-platform-admin-ui/commit/1d881a70ddfed165ba85017d517f56e9e7b2c490))

## 1.0.0 (2023-07-07)


### Features

* Add go code skeleton ([10aec9d](https://github.com/canonical/identity-platform-admin-ui/commit/10aec9d8f2181d7c6c0d5cc2aebf861381827139))
* add ui skeleton ([ce6b51f](https://github.com/canonical/identity-platform-admin-ui/commit/ce6b51ff0659c16751b7d2371d4b19f399cad59a))
