# Changelog

## [1.8.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.7.0...v1.8.0) (2024-05-09)


### Features

* upgrade rebac-admin to 0.0.1-alpha.3 ([96aca77](https://github.com/canonical/identity-platform-admin-ui/commit/96aca771bfe328bc27e8e656bd471345a5c43b25))

## [1.7.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.6.1...v1.7.0) (2024-05-06)


### Features

* implement new Create{Group,Role} interface + adjust handlers ([0adce3c](https://github.com/canonical/identity-platform-admin-ui/commit/0adce3cd75227c7cfa1d479ca94b69ef8eea6b86))
* let Create{Group,Role} return newly created object ([e1ba968](https://github.com/canonical/identity-platform-admin-ui/commit/e1ba96806f98fdf05d56104f44632f6bb935c274))

## [1.6.1](https://github.com/canonical/identity-platform-admin-ui/compare/v1.6.0...v1.6.1) (2024-05-06)


### Bug Fixes

* **role:** error out when ID is passed for creation ([2a46a5e](https://github.com/canonical/identity-platform-admin-ui/commit/2a46a5ec7a04a3aafbe6a899167acfe8ea8ff00e))
* **role:** use `Name` field for creation ([e63fdaa](https://github.com/canonical/identity-platform-admin-ui/commit/e63fdaa70fb58e8c3001a23cc2ecc9e140e9cb81))

## [1.6.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.5.0...v1.6.0) (2024-04-30)


### Features

* add `openfga_workers_total` int config with default ([b12ac05](https://github.com/canonical/identity-platform-admin-ui/commit/b12ac05c0a655932a95ca7384d1cbd9d995d238b))
* add `payload_validation_enabled` config key ([419b042](https://github.com/canonical/identity-platform-admin-ui/commit/419b042e22fe2d741afec7acfe0d54c10889b07d))
* add `SetTokens` method + empty tokens don't get set ([f165155](https://github.com/canonical/identity-platform-admin-ui/commit/f16515588bea0125e7e03d9f5b0f058a96970254))
* add 3rd party validator to API structs + setupValidation func + initial noop middleware ([1de0006](https://github.com/canonical/identity-platform-admin-ui/commit/1de0006c1db9b7d7f32c79f13de797420427db2b))
* add constructor for validator + use json tags for validation errors ([44d7223](https://github.com/canonical/identity-platform-admin-ui/commit/44d7223b6466d5cab9fadf851d4830e7d8ae0062))
* add externalized Kube config file env var ([9a63fe3](https://github.com/canonical/identity-platform-admin-ui/commit/9a63fe3f544784d4a420b88ed3529a1514d9e7bd))
* add full validation implementation for schemas ([45993ed](https://github.com/canonical/identity-platform-admin-ui/commit/45993ed14506cd90f9f019d5317b4df29d726e22))
* add identity provider management, add logo ([48f47ec](https://github.com/canonical/identity-platform-admin-ui/commit/48f47ec41daf2cd09304f5745b462e9795af6540))
* add log tailing to skaffold run ([a9725da](https://github.com/canonical/identity-platform-admin-ui/commit/a9725da88b358cc487e84117de82ba4e98ee38ae))
* add login screen ([1befe87](https://github.com/canonical/identity-platform-admin-ui/commit/1befe87ce968dc0b4c9badc6a8b8543d3b281096))
* add pagination to clients, schemas and identity lists in ui. Add identity creation form WD-10253 ([5f55463](https://github.com/canonical/identity-platform-admin-ui/commit/5f554639a669404b5e468fd93c77af9e52cd946b))
* add URL param validation for groups handlers ([24c8d99](https://github.com/canonical/identity-platform-admin-ui/commit/24c8d99319e1782cd742451d9b09f6846bd6fa3e))
* add Urn type ([f7d33e2](https://github.com/canonical/identity-platform-admin-ui/commit/f7d33e2ab27411aeb4ce82ace2ab345cc45c6888))
* add validation implementation for `clients` ([549d985](https://github.com/canonical/identity-platform-admin-ui/commit/549d985ed5ded7f8b1522479208d1637bb5e6855))
* add validation implementation for `groups` ([700cf04](https://github.com/canonical/identity-platform-admin-ui/commit/700cf0401d657a771e56511bd04f95cea93675e6))
* add validation middlewareonly if payload validation is enabled + reorder middleware and endpoints registration ([32814e8](https://github.com/canonical/identity-platform-admin-ui/commit/32814e89103c5abfc8be4a144e4343f26ff85012))
* add validation setup for `groups` endpoint ([06fb9f4](https://github.com/canonical/identity-platform-admin-ui/commit/06fb9f4c777b880b4be1fb646360e9cf6b805095))
* add validation setup for `identities` endpoint ([b4178c9](https://github.com/canonical/identity-platform-admin-ui/commit/b4178c95c2771b2149fb92cc80d43431b6c7028b))
* add validation setup for `schemas` endpoint ([8c5e173](https://github.com/canonical/identity-platform-admin-ui/commit/8c5e17319243cc44dbe3d353acb2df57819334ac))
* add ValidationRegistry for API validation + instantiate in router ([50f0810](https://github.com/canonical/identity-platform-admin-ui/commit/50f08107ceee72ec40d05f0477e4898bd70b3347))
* add worker pool implementation ([dbd2f9d](https://github.com/canonical/identity-platform-admin-ui/commit/dbd2f9d74e3b0045f6475ec112ae18c444ae62d5))
* adjust identity api to accept page token ([beb0d42](https://github.com/canonical/identity-platform-admin-ui/commit/beb0d429af14d494b1d5edbe8598460acf4c4685)), closes [#256](https://github.com/canonical/identity-platform-admin-ui/issues/256)
* adjust pagination for schemas endpoints ([e2a2df3](https://github.com/canonical/identity-platform-admin-ui/commit/e2a2df3c57e02377dd159e022e6de34fc44e1780)), closes [#44](https://github.com/canonical/identity-platform-admin-ui/issues/44)
* allow create-fga-model cli command to save on a k8s coonfigmap ([56463bb](https://github.com/canonical/identity-platform-admin-ui/commit/56463bb2db0759ef14c876177a7087fdecc463fe))
* authorization middleware based on openFGA ([8f2cb3e](https://github.com/canonical/identity-platform-admin-ui/commit/8f2cb3e4b0723d531704d2c68f4bbe6d07851efd))
* create groups service ([3d8d648](https://github.com/canonical/identity-platform-admin-ui/commit/3d8d648081d2629d6a7c360a0d2934fdc5e3d438))
* create roles service ([c796135](https://github.com/canonical/identity-platform-admin-ui/commit/c796135b8557998d05c72f4295948b4f8c15403e))
* create token pagination extractor ([215b6cb](https://github.com/canonical/identity-platform-admin-ui/commit/215b6cbd8c1e34a80c072a9210e4e48d2df875aa))
* **create-group:** allow creator user to view group ([efcaeec](https://github.com/canonical/identity-platform-admin-ui/commit/efcaeecc079040b02f89a4b87a8e1fe48e709076))
* **delete-group:** delete all relation for group to delete ([883b513](https://github.com/canonical/identity-platform-admin-ui/commit/883b513909d0deadf4e7027f4c5f7f1ef998b5c8))
* enable authorization by default ([6f61651](https://github.com/canonical/identity-platform-admin-ui/commit/6f616518b08b761002dfb6a229aa9a0b5098e713))
* enhance identity provider form to cover all providers and relevant fields, hide advanced fields by default ([ef62667](https://github.com/canonical/identity-platform-admin-ui/commit/ef626673a0cb7ea767395531785892c19c4273dc))
* enhance ValidationRegistry with PayloadValidator and adjust in handlers + enhance Middleware + add func for ApiKey retrieval from endpoint ([313617a](https://github.com/canonical/identity-platform-admin-ui/commit/313617a7faaf8292df5b0a5cfc509f9e40188290))
* enhanced ValidationError with specific field errors and common errors ([a21462c](https://github.com/canonical/identity-platform-admin-ui/commit/a21462c78249d83961ad19a167ceeb57e5366e1f))
* handlers for groups API ([63d5dc4](https://github.com/canonical/identity-platform-admin-ui/commit/63d5dc4bcfef3a909a942a30d5f486d23209a4ed))
* handlers for roles API ([114b284](https://github.com/canonical/identity-platform-admin-ui/commit/114b284fd3a205ebb4879b61c440e5cedc51c9db))
* hook up worker pool for groups and roles API ([ce83bd6](https://github.com/canonical/identity-platform-admin-ui/commit/ce83bd6a1649caf67eef42b42a322ecb178fdece))
* **idp:** add validation implementation ([71ff661](https://github.com/canonical/identity-platform-admin-ui/commit/71ff6612485dd73374e09508143d50f455a46270))
* implement converters for each type of API ([09852b0](https://github.com/canonical/identity-platform-admin-ui/commit/09852b03626a05e9034bfe3641b0ca667801d992))
* include roles and groups from ReBAC Admin ([5d03914](https://github.com/canonical/identity-platform-admin-ui/commit/5d03914cd12732584d37f0d0e31c5668ce960c25))
* introduce BatchCheck, WriteTuples, DeleteTuples and ReadTuples in openfga client ([39eb195](https://github.com/canonical/identity-platform-admin-ui/commit/39eb195e4adcf9a05339d3126f44a1f3bf805e6e))
* introduce groups API converter to deal with authorization in the middleware ([5f8875a](https://github.com/canonical/identity-platform-admin-ui/commit/5f8875aa26a1d5fab0c6a0f115d3d1ab17a8b7a9))
* invoke setup validation on registered APIs ([de16a0b](https://github.com/canonical/identity-platform-admin-ui/commit/de16a0bc7829bf1a849c1f06b408408e0845e365))
* parse and expose link header from hydra ([7c2d3f6](https://github.com/canonical/identity-platform-admin-ui/commit/7c2d3f656f57e0594f890656df34f941fd0fce78))
* passing openfga store and model id to admin service ([51f4fab](https://github.com/canonical/identity-platform-admin-ui/commit/51f4fab77a70c9a77a1661f88d64b5e0865a9c5e))
* **roles:** add validation implementation ([6bf72e5](https://github.com/canonical/identity-platform-admin-ui/commit/6bf72e5d75d94daca1fdf028ed1a3f7744e67b4b))
* **rules:** add validation implementation ([c42bd45](https://github.com/canonical/identity-platform-admin-ui/commit/c42bd45cf7af8b1fe46858c8480693bda8dc9145))
* separate authorization client from OpenFGA client ([2cc4dab](https://github.com/canonical/identity-platform-admin-ui/commit/2cc4dabb6a9f75b558fea627c4d4c4bed783b472))
* upgrade openfga model ([c49abd5](https://github.com/canonical/identity-platform-admin-ui/commit/c49abd55aa5e85a59f9c030b2e9bc032fa38b21c))
* use interface instead of client pointer ([3e1ac0f](https://github.com/canonical/identity-platform-admin-ui/commit/3e1ac0f9ebcb8b460a661e1e4506fea687973aff))
* use side panels for client and idp creation ([ef798c4](https://github.com/canonical/identity-platform-admin-ui/commit/ef798c4a0d177cc0abd6cb0d6bd1ee0aecc8fb64))
* wire up groups API ([352bc45](https://github.com/canonical/identity-platform-admin-ui/commit/352bc45665936ba70f180990839fd70df590ce3c))
* wire up roles API in web application ([16ba352](https://github.com/canonical/identity-platform-admin-ui/commit/16ba3521f18a18b233a6cea84eccf687952d1890))


### Bug Fixes

* adapt serve command to changes on k8s client ([e6701e2](https://github.com/canonical/identity-platform-admin-ui/commit/e6701e22ccc319fc3f4e17957829f6111b245d18))
* add back URL Param validation from previous commit ([ebe07a5](https://github.com/canonical/identity-platform-admin-ui/commit/ebe07a5d9b2badcdeb4616a0ccd5d753374fedac))
* add command for creating an admin user ([50449a9](https://github.com/canonical/identity-platform-admin-ui/commit/50449a9e43f9a886f181014c3cbb8c8b9c576a5c))
* add command for removing an admin user ([2db3a08](https://github.com/canonical/identity-platform-admin-ui/commit/2db3a0885f35043d2963a194c5585f42ca94c172))
* add extra check on list schemas test for navigation ([2afec86](https://github.com/canonical/identity-platform-admin-ui/commit/2afec86c79e20912490e6bbcf5e3218f961d5b29))
* add id validation to make sure it's never empty ([fc7d560](https://github.com/canonical/identity-platform-admin-ui/commit/fc7d5606988a05a668a3a51c7458a2b32a4a0042)), closes [#239](https://github.com/canonical/identity-platform-admin-ui/issues/239)
* add page tokens to the response ([5a13e4e](https://github.com/canonical/identity-platform-admin-ui/commit/5a13e4e1105fe085230ac85d1fea127ea9ba8f23))
* add todo comment to catch issue with the user-identities sync ([ed66418](https://github.com/canonical/identity-platform-admin-ui/commit/ed66418499058ff26605b07f8b196a3d3ba2ab6d))
* add validation to openfga config ([300201c](https://github.com/canonical/identity-platform-admin-ui/commit/300201ccce5fdd767d3918377976634ee9f6ae28))
* address empty IDs on schema and idp creation ([e6dbf32](https://github.com/canonical/identity-platform-admin-ui/commit/e6dbf32c94e92ce5d79e9b1cb383c8f1243b943c)), closes [#227](https://github.com/canonical/identity-platform-admin-ui/issues/227)
* address segfault when using noop client ([5265512](https://github.com/canonical/identity-platform-admin-ui/commit/5265512f773bc5e8432c68a972e4a7a123f0075c))
* adjust openfga NoopClient setup ([f253400](https://github.com/canonical/identity-platform-admin-ui/commit/f253400882b9b4a3809b1ca4aa468751705f6c2a))
* adjust page offset for oathkeeper apis ([7c22e06](https://github.com/canonical/identity-platform-admin-ui/commit/7c22e065f12503625a78b2a4e33f19314aaa376c))
* allow for k8s client to be configured using kubeconfig ([136e957](https://github.com/canonical/identity-platform-admin-ui/commit/136e9572f81485103baf12d66f2eed9b61657661))
* bundle up external clients and o11y setup into config structs ([a660066](https://github.com/canonical/identity-platform-admin-ui/commit/a660066d58cce54b8e18a2968a9a6ce7bc0cd25d))
* change specs.EnvSper name for authorization model id ([3eb270b](https://github.com/canonical/identity-platform-admin-ui/commit/3eb270b01f7acbea330672eb66e07c8f8f2e3ba3))
* **clients:** validation and improved tests ([129a8a8](https://github.com/canonical/identity-platform-admin-ui/commit/129a8a8b40ae33cf2f531fae721c17837f12cb7e))
* deal with empty Data attribute in k8s configmap ([56937c8](https://github.com/canonical/identity-platform-admin-ui/commit/56937c87497e40b03e881af4ab5019595e3a6f55)), closes [#254](https://github.com/canonical/identity-platform-admin-ui/issues/254)
* delete role implementation ([4b71734](https://github.com/canonical/identity-platform-admin-ui/commit/4b717346d354c54c31242cedfa7461a049faf0d7))
* disable validation due to missing implementation of api validators ([5c06b9b](https://github.com/canonical/identity-platform-admin-ui/commit/5c06b9b540a881fa21eb03ecd07fa810ee5a7693))
* drop non can_ relations from group entitlements ([5b225ae](https://github.com/canonical/identity-platform-admin-ui/commit/5b225aecc13ce1f51d7c848a7c7f58dd25ad7843)), closes [#243](https://github.com/canonical/identity-platform-admin-ui/issues/243)
* enhance cli model creation to bootstrap a store ([e97fb0a](https://github.com/canonical/identity-platform-admin-ui/commit/e97fb0afefe191813c5c58077a2924f5df9b1f2c))
* enhance openfga client with CreateStore and helpers to set modelID and storeID on the fly ([5d62fbf](https://github.com/canonical/identity-platform-admin-ui/commit/5d62fbf7c122adceb99b228f850714571c95034b))
* enhance registerValidation log message with error ([ae95fa8](https://github.com/canonical/identity-platform-admin-ui/commit/ae95fa8264cc1df58e69f2f50124b79f0fd4a354))
* get 404 with not found role (with can view) - get 403 (without can_view) ([2a22054](https://github.com/canonical/identity-platform-admin-ui/commit/2a22054c2cc1c63128dc5a75f050e4bf5df6c8d1))
* **groups:** validation and improved tests ([255733e](https://github.com/canonical/identity-platform-admin-ui/commit/255733e3d5499181c2ef9b92f9145ae7997541ce))
* handleDetail to return 404 on missing group for authorized users + typo ([b1a1e02](https://github.com/canonical/identity-platform-admin-ui/commit/b1a1e0222a5ba2f1d2c3c26e4fe566c1877f4dcd))
* **identities:** validation and improved tests ([b4fa762](https://github.com/canonical/identity-platform-admin-ui/commit/b4fa7629306681e25b16c9d7cadcfdcd96fdef02))
* introduce uri validation for params ([5eecee4](https://github.com/canonical/identity-platform-admin-ui/commit/5eecee4ac5f72c2d8a536c812a8468bf3bd86000))
* listing not working for user that created a role ([b54d681](https://github.com/canonical/identity-platform-admin-ui/commit/b54d6811be1d6f041cad64fa15c0586eec530f35))
* pass interface to roles API to allow for openfga noop client ([6d04a3d](https://github.com/canonical/identity-platform-admin-ui/commit/6d04a3d689c79d2b87831c770a4dbbdb9feb7a75))
* remove assignees tuples on DeleteGroup ([1107165](https://github.com/canonical/identity-platform-admin-ui/commit/1107165dc59998915a88b4e5ad7ec35db53161ee))
* remove assignees tuples on DeleteRole ([5772334](https://github.com/canonical/identity-platform-admin-ui/commit/57723345d44f50e5faac89547779cafe5c644dab)), closes [#285](https://github.com/canonical/identity-platform-admin-ui/issues/285)
* remove page_token field in meta response ([3756f0d](https://github.com/canonical/identity-platform-admin-ui/commit/3756f0d7493afafe4649da22cce814fb7acf8952)), closes [#271](https://github.com/canonical/identity-platform-admin-ui/issues/271)
* removing extra #member on assignIdentities service call ([bfde070](https://github.com/canonical/identity-platform-admin-ui/commit/bfde070133a505f5f38b89ec19c12e4378c63ca9)), closes [#283](https://github.com/canonical/identity-platform-admin-ui/issues/283)
* removing extra #member on removeIdentities service call ([74ab0ff](https://github.com/canonical/identity-platform-admin-ui/commit/74ab0fff68c20196387a0a32c5226110ffcf6ed5))
* **schemas:** validation and improved tests ([ab8652f](https://github.com/canonical/identity-platform-admin-ui/commit/ab8652f13c437cf64ca6978726b252059f4fb324))
* skip validation config on createFGAmodel cmd ([ffd6563](https://github.com/canonical/identity-platform-admin-ui/commit/ffd6563ddeaef17d7041908e1184c2fd0bcaebb5))
* standardize on types.Response ([02cc8ce](https://github.com/canonical/identity-platform-admin-ui/commit/02cc8ceafe338bd75910bb307415af668d1d1761)), closes [#244](https://github.com/canonical/identity-platform-admin-ui/issues/244)
* standardize page token in clients api ([7bdd3e7](https://github.com/canonical/identity-platform-admin-ui/commit/7bdd3e7b61bda2675f757861195af87d063e59db))
* switch to use WriteTuples instead of WriteTuple ([ba8a624](https://github.com/canonical/identity-platform-admin-ui/commit/ba8a624f57af16ebea3889b77baf9260d2589ad6))
* update noop openfga client with newer methods ([251a8a1](https://github.com/canonical/identity-platform-admin-ui/commit/251a8a1b0be4935063f9e49927b06d8291c7d985))
* use sync.Map for race conditions ([603a7e1](https://github.com/canonical/identity-platform-admin-ui/commit/603a7e1fa80aec118375a30a3d73e5b124847103))
* use the microk8s-hostpath storageclass to dynamically provision the persistent volume ([29d8f39](https://github.com/canonical/identity-platform-admin-ui/commit/29d8f39f50f8951b56b17a5f5fc69765e092f81f))
* wire up new config structs into web application bootstrap ([9e5587d](https://github.com/canonical/identity-platform-admin-ui/commit/9e5587d0cfc0e87228c57bef0892c438c5adf07b)), closes [#222](https://github.com/canonical/identity-platform-admin-ui/issues/222)

## [1.5.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.4.0...v1.5.0) (2024-01-26)


### Features

* use cobra-cli ([8f061d3](https://github.com/canonical/identity-platform-admin-ui/commit/8f061d3168545cb8d5a04770cd06549a29ba3c2d))


### Bug Fixes

* add config for openfga integration ([bc751e2](https://github.com/canonical/identity-platform-admin-ui/commit/bc751e2842b6791c96a9316d26ebbb3d0e500944))
* add logic for create-fga-model ([7fc9a6c](https://github.com/canonical/identity-platform-admin-ui/commit/7fc9a6c4bba714cfba23e6a880e3314c9fc68126))
* add noop tracer ([f97484c](https://github.com/canonical/identity-platform-admin-ui/commit/f97484cf2758f66506c93262214152898099f08a))
* add openfga module ([d7d3418](https://github.com/canonical/identity-platform-admin-ui/commit/d7d34183d7ec1c9513d08577e26fcb1606d4cae2))
* implement version command ([fe5fc83](https://github.com/canonical/identity-platform-admin-ui/commit/fe5fc8359cbb06d3e4921d803240b4954def2fc0))
* introduce authorization module ([28df12b](https://github.com/canonical/identity-platform-admin-ui/commit/28df12bd29de1f2bd085a52f9aa012e517cb7325))
* introduce noop logging and monitoring ([09b529d](https://github.com/canonical/identity-platform-admin-ui/commit/09b529d3a519a30bdb46c5fe18ea5c782429cdd5))

## [1.4.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.3.0...v1.4.0) (2024-01-04)


### Features

* added unit tests for pkg/rules package ([e36bbd3](https://github.com/canonical/identity-platform-admin-ui/commit/e36bbd3aa2b95e8bb9ce0f4cd57c42c4c16e7c8a))
* implemented interface for manipulating Oathkeeper rules ([e36bbd3](https://github.com/canonical/identity-platform-admin-ui/commit/e36bbd3aa2b95e8bb9ce0f4cd57c42c4c16e7c8a))


### Bug Fixes

* fixed issue with make dev ([0d81544](https://github.com/canonical/identity-platform-admin-ui/commit/0d81544849133ea268e55c3338b4131d0a2e61b4))
* fixed issues with make dev ([0d81544](https://github.com/canonical/identity-platform-admin-ui/commit/0d81544849133ea268e55c3338b4131d0a2e61b4))
* make rules cm file name configurable ([3f05b59](https://github.com/canonical/identity-platform-admin-ui/commit/3f05b59bbe7571d7bc1f0d99c72ff5381e9aa54d))

## [1.3.0](https://github.com/canonical/identity-platform-admin-ui/compare/v1.2.0...v1.3.0) (2023-11-03)


### Features

* add schemas endpoints ([c9be3dc](https://github.com/canonical/identity-platform-admin-ui/commit/c9be3dcf364b62c9900233f56cc448d51a5f3cf7))
* add schemas service layer and interfaces ([83917cf](https://github.com/canonical/identity-platform-admin-ui/commit/83917cf291f031a01d8882364ac3e50aedfd99e5))
* add unit tests for default schema feature ([777259a](https://github.com/canonical/identity-platform-admin-ui/commit/777259ad37849b6716edd0955708c4656f92a964))
* added ca-certificates package to stage-packages ([16f6683](https://github.com/canonical/identity-platform-admin-ui/commit/16f6683218d02e3c60923007a59cf36c2bb5f5d2))
* wire up schemas pkg ([513ce61](https://github.com/canonical/identity-platform-admin-ui/commit/513ce612809910c78bfc4f4647246ce4f90a1c42))


### Bug Fixes

* add default schema changes ([82ba9d6](https://github.com/canonical/identity-platform-admin-ui/commit/82ba9d6fb1bdee1a62d3ee8243581a7a6431804e))
* **deps:** update dependency @canonical/react-components to v0.47.0 ([#94](https://github.com/canonical/identity-platform-admin-ui/issues/94)) ([a2c7e03](https://github.com/canonical/identity-platform-admin-ui/commit/a2c7e0318bbeb20d3cb7aaf86122b8dd1ada49fc))
* **deps:** update dependency @canonical/react-components to v0.47.1 ([7b6cec0](https://github.com/canonical/identity-platform-admin-ui/commit/7b6cec025e2be7a399c439015f0e5287082ec20d))
* **deps:** update dependency sass-embedded to v1.67.0 ([#106](https://github.com/canonical/identity-platform-admin-ui/issues/106)) ([4a5922c](https://github.com/canonical/identity-platform-admin-ui/commit/4a5922c5e220f42717dbdac61dbf56568ba604db))
* **deps:** update dependency sass-embedded to v1.69.1 ([#137](https://github.com/canonical/identity-platform-admin-ui/issues/137)) ([3bc1132](https://github.com/canonical/identity-platform-admin-ui/commit/3bc113283b9523bc6202a23527e01a8a9f98345c))
* **deps:** update dependency sass-embedded to v1.69.2 ([#141](https://github.com/canonical/identity-platform-admin-ui/issues/141)) ([1533b21](https://github.com/canonical/identity-platform-admin-ui/commit/1533b21c56566fb3ecc8f863d26773ac3a04ebb6))
* **deps:** update dependency sass-embedded to v1.69.4 ([d695e33](https://github.com/canonical/identity-platform-admin-ui/commit/d695e33c690ed2365d348b3a45870407e399ff34))
* **deps:** update dependency vanilla-framework to v4 ([#95](https://github.com/canonical/identity-platform-admin-ui/issues/95)) ([35c21ae](https://github.com/canonical/identity-platform-admin-ui/commit/35c21aea82d6c6b4cb1501d771bd249d44d476e4))
* **deps:** update dependency vanilla-framework to v4.3.0 ([#99](https://github.com/canonical/identity-platform-admin-ui/issues/99)) ([049629c](https://github.com/canonical/identity-platform-admin-ui/commit/049629c98e36c433e144595e6b4ad25f1f6872d9))
* **deps:** update dependency vanilla-framework to v4.4.0 ([dde2c11](https://github.com/canonical/identity-platform-admin-ui/commit/dde2c1122a1975a0586332ff2c89d9704451da98))
* **deps:** update dependency vanilla-framework to v4.5.0 ([b700785](https://github.com/canonical/identity-platform-admin-ui/commit/b7007852193a600ae054c91a8f7708d761402b21))
* **deps:** update go deps (minor) ([#101](https://github.com/canonical/identity-platform-admin-ui/issues/101)) ([2f1e289](https://github.com/canonical/identity-platform-admin-ui/commit/2f1e289aabf71945582603b64dd8cc8561e125e1))
* **deps:** update go deps (minor) ([#127](https://github.com/canonical/identity-platform-admin-ui/issues/127)) ([903ee82](https://github.com/canonical/identity-platform-admin-ui/commit/903ee827dd8d4a744abd52fe508b9e4c0f5e32ae))
* **deps:** update go deps (minor) ([#75](https://github.com/canonical/identity-platform-admin-ui/issues/75)) ([54f9421](https://github.com/canonical/identity-platform-admin-ui/commit/54f9421d686543e552e04d9790843db83d90c103))
* **deps:** update go deps to v0.28.2 (patch) ([#105](https://github.com/canonical/identity-platform-admin-ui/issues/105)) ([5888133](https://github.com/canonical/identity-platform-admin-ui/commit/5888133c9709899ca389296214101ce534b0061f))
* **deps:** update go deps to v0.28.3 ([10422e3](https://github.com/canonical/identity-platform-admin-ui/commit/10422e3c96924a961f11a2ac661769a166e906c5))
* **deps:** update go deps to v1.17.0 (minor) ([#71](https://github.com/canonical/identity-platform-admin-ui/issues/71)) ([472dc50](https://github.com/canonical/identity-platform-admin-ui/commit/472dc5067964d4fc183c677d82cf3a791d646cd0))
* **deps:** update go deps to v1.18.0 (minor) ([#100](https://github.com/canonical/identity-platform-admin-ui/issues/100)) ([129c7ee](https://github.com/canonical/identity-platform-admin-ui/commit/129c7eeedb143af2422063f20ae83b066543fcba))
* **deps:** update go deps to v1.19.0 (minor) ([#125](https://github.com/canonical/identity-platform-admin-ui/issues/125)) ([1d870ba](https://github.com/canonical/identity-platform-admin-ui/commit/1d870ba7f0a25a55fc7d70b7f1594f72f6e0df33))
* **deps:** update module github.com/google/uuid to v1.3.1 ([#53](https://github.com/canonical/identity-platform-admin-ui/issues/53)) ([840b068](https://github.com/canonical/identity-platform-admin-ui/commit/840b0689e8b75c3bd001e8804a1f2bec471ec47d))
* **deps:** update module github.com/google/uuid to v1.4.0 ([2ce70cf](https://github.com/canonical/identity-platform-admin-ui/commit/2ce70cf126f06d9d5092936cf6c87d4bbf52362b))
* **deps:** update module github.com/ory/kratos-client-go to v1 ([4fefc13](https://github.com/canonical/identity-platform-admin-ui/commit/4fefc13cb30cd37d3090afb27185fe60c00a6859))
* **deps:** update module github.com/prometheus/client_golang to v1.17.0 ([#124](https://github.com/canonical/identity-platform-admin-ui/issues/124)) ([e0904d9](https://github.com/canonical/identity-platform-admin-ui/commit/e0904d9e2fe6cdfc58feb0678a6057ff6e71eea2))
* **deps:** update module go.opentelemetry.io/otel/exporters/stdout/stdouttrace to v1.17.0 ([#72](https://github.com/canonical/identity-platform-admin-ui/issues/72)) ([9fd027b](https://github.com/canonical/identity-platform-admin-ui/commit/9fd027b2a1818676b973882e67009bf494a01cd6))
* **deps:** update module go.uber.org/zap to v1.26.0 ([#111](https://github.com/canonical/identity-platform-admin-ui/issues/111)) ([f836ac3](https://github.com/canonical/identity-platform-admin-ui/commit/f836ac342253fdf2b3655a9ee02972108b36e6f6))
* fix renovate config ([700cc51](https://github.com/canonical/identity-platform-admin-ui/commit/700cc515c2af7a56d5e5781ecd38de0cb29aaaa4))
* fixed struct inconsistencies with the new release of kratos-client-go ([3808420](https://github.com/canonical/identity-platform-admin-ui/commit/38084207eb6867026a592aa3cff00945f0fcdd97))
* introduce version flag to facilitate charm code ([4a1b6e1](https://github.com/canonical/identity-platform-admin-ui/commit/4a1b6e1fdec7c9a673a31f82e3ae77ff47f76f2f))
* use version from release-please worflow ([450c0bd](https://github.com/canonical/identity-platform-admin-ui/commit/450c0bd686d91ad4d59ae7f01ca8bff9f71c0a16))
* use version in /api/v0/version endpoint ([cdc9297](https://github.com/canonical/identity-platform-admin-ui/commit/cdc9297e3a913a0c3abfd6b5cc0c4d880edef5fa))

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
