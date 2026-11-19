# Changelog

## [0.2.0](https://github.com/Doratou/Backend/compare/v0.1.0...v0.2.0) (2026-11-19)


### ⚠ BREAKING CHANGES

* **settlements:** Settlement responses no longer include cancel_reason.
* **settlements:** Settlement responses no longer include rejection_reason.
* **polls:** support single and multiple choice voting
* **draws:** Lucky draw creation now requires draw_type and typed type/value candidates. WEIGHTED chances must total 100 percent. Groups share one draft that any active member can execute.

### Features

* add weighted lucky draws and complete poll list responses ([edaea26](https://github.com/Doratou/Backend/commit/edaea26dd5b31a9074c3b7fdd1aad1bfb8a8ac17))
* add weighted lucky draws and complete poll list responses ([#18](https://github.com/Doratou/Backend/issues/18)) ([edaea26](https://github.com/Doratou/Backend/commit/edaea26dd5b31a9074c3b7fdd1aad1bfb8a8ac17))
* **ai:** add ai service and enum and exception ([71109c0](https://github.com/Doratou/Backend/commit/71109c0652ec2b6db2e397ef2494646c29eea0af))
* **ai:** save AI response to database and weather API to cache and add list suggestions endpoint ([29a154d](https://github.com/Doratou/Backend/commit/29a154d737983628dbc785766a421770041dfe5d))
* **api:** prune redundant routes, add avatar/cover endpoints, and wire optional auth middleware ([b9ca350](https://github.com/Doratou/Backend/commit/b9ca3508e0eceed8a8cf15fa76d14143eda15357))
* **attachment:** establish relations with groups and expenses and add base expense schema ([9eda188](https://github.com/Doratou/Backend/commit/9eda188d4023fa195196bdbea45bcc6a92bf2d78))
* **attachment:** expose group and expense query endpoints and relation hydration ([0a3007e](https://github.com/Doratou/Backend/commit/0a3007e5ff97849fd7d4e1ecd8c7e3287310b31b))
* **attachment:** guard permanent deletion endpoint with admin RBAC ([b4fede0](https://github.com/Doratou/Backend/commit/b4fede05f04782c02f61234fbf68ee02f344b5ae))
* **auth:** implement Google OAuth2/OIDC authentication and account linking ([9518347](https://github.com/Doratou/Backend/commit/95183474f3492739255362a8b2dc13147dededcf))
* **ci:** added build-docs job ([5df8cfa](https://github.com/Doratou/Backend/commit/5df8cfa509eecaeb4f4a367ccbe104e287eae5ad))
* **ci:** added Dockerfile for docs ([b3af4df](https://github.com/Doratou/Backend/commit/b3af4dfea0dd6a5b6897e870432680d6c1c99461))
* **config,exception:** PoW two-layer rate limiter config and error types ([df0edf5](https://github.com/Doratou/Backend/commit/df0edf537b7521ea5baa578350a4c50820715dd4))
* **core:** implement context guards, repo unclaimed query, and service methods for attachments, auth, and groups ([9219f43](https://github.com/Doratou/Backend/commit/9219f431015e252a8cb8a21bf022d36c60ae76ae))
* **draws:** add domain model, errors and repository contract ([0fb4dcd](https://github.com/Doratou/Backend/commit/0fb4dcd781b5697c4e9c06dd6930512cab97c8db))
* **draws:** add HTTP handlers with validation and standard error responses ([781db0e](https://github.com/Doratou/Backend/commit/781db0e02de615eee96de55e45e7930630cccf45))
* **draws:** add typed candidates and shared percentage-based draws ([9ff73c8](https://github.com/Doratou/Backend/commit/9ff73c82c2022e0c5f34555a424d2f5d6a0971e6))
* **draws:** add weighted lucky draw migration ([e4051ef](https://github.com/Doratou/Backend/commit/e4051eff0f5a5d877d42cd50044f3cd711d2e828))
* **draws:** implement PostgreSQL repository and secure random generator ([12b8ec5](https://github.com/Doratou/Backend/commit/12b8ec5ddb550d8d3bbb94f0e4e156dc3dc444ba))
* **draws:** implement weighted draw service with validation and access control ([cdc50ad](https://github.com/Doratou/Backend/commit/cdc50ad07b38d51269455fd35bd080593784371c))
* **draws:** register authenticated routes and wire application dependencies ([00113f8](https://github.com/Doratou/Backend/commit/00113f825f4ded85f125c222e86257f62f14e65b))
* **enum:** add genenum generator for automated domain enum helpers ([1017e4f](https://github.com/Doratou/Backend/commit/1017e4fb4a551b6f89c5a74c347297a6af4e228f))
* **expenses:** add expense handlers and enforce attachment access checks ([f2adc1a](https://github.com/Doratou/Backend/commit/f2adc1aa6bfdeb9d57df7217c9536a33591b62b8))
* **expenses:** add paginated reports for group, created and owed expenses ([fb89471](https://github.com/Doratou/Backend/commit/fb8947162dd5c6ca026a21e69d80dcc50ac62b46))
* **expenses:** add split types and expense shares migration ([5a81b7b](https://github.com/Doratou/Backend/commit/5a81b7b3d58ebd7ba7e7a134b55a7a428727a07a))
* **expenses:** define expense entities, split enums and repository contracts ([d939506](https://github.com/Doratou/Backend/commit/d93950664f6aeebc31509612aca2c36bba929696))
* **expenses:** implement expense services with split and receipt validation ([2defae6](https://github.com/Doratou/Backend/commit/2defae694796f130d5b2255a2fba51b54122b32b))
* **expenses:** implement PostgreSQL repositories for expenses, shares and receipts ([69aeaa3](https://github.com/Doratou/Backend/commit/69aeaa3f016bab2ee503083fd85b91ed8bf81007))
* **expenses:** merge shared expense management into develop ([6994dc0](https://github.com/Doratou/Backend/commit/6994dc06774f0a5cc9abf92f1abdf9f17889d02c))
* **expenses:** register expense routes and wire application dependencies ([6e71a48](https://github.com/Doratou/Backend/commit/6e71a482a9dfdbbaeda3e555061b75b5b096c588))
* **friendships:** add avatar URLs and mutual group counts to lists ([95c15be](https://github.com/Doratou/Backend/commit/95c15be881acb83cfb33a3e892ac3223c4687673))
* **friendships:** merge friend list details into develop ([5b7559c](https://github.com/Doratou/Backend/commit/5b7559c133872dfb2de9d3d14c8e9609054497cf))
* **group:** add group seed data ([f6a7123](https://github.com/Doratou/Backend/commit/f6a7123b01d5fceaeb1b5a1443052427475866b4))
* **hamyar:** add hamyar handler and router ([68f2010](https://github.com/Doratou/Backend/commit/68f2010fb9db140a06b35b5b07adab0c25639f99))
* **middleware:** TwoLayerLimiter with stateless PoW challenge flow ([7cb0b1a](https://github.com/Doratou/Backend/commit/7cb0b1a329b8ba8a142e31f75afb869317d8bdda))
* **polls:** include options, vote counts and user choice in poll lists ([5c24fc2](https://github.com/Doratou/Backend/commit/5c24fc2b2ed52a058ea5fe414c25723549283dd9))
* **polls:** support single and multiple choice voting ([a01a47b](https://github.com/Doratou/Backend/commit/a01a47b460c7471b8589d19b7b868efb19cad6fb))
* **pow,config,repository:** two-layer rate limiter foundation with stateless PoW challenges ([27a632f](https://github.com/Doratou/Backend/commit/27a632f3712d3445ca6fba1d57007ac39788cbe9))
* **pow:** stateless HMAC-signed challenge generation and verification ([0868002](https://github.com/Doratou/Backend/commit/086800293de5b0311cdfa6b89ac276592c35355e))
* **repository:** PoW block repository interface and Redis implementation ([a2250dd](https://github.com/Doratou/Backend/commit/a2250ddca2214edc29b227cc60e1927f0c285b0a))
* **security:** implement RBAC middleware and guard admin routes for reports and groups ([a4795b3](https://github.com/Doratou/Backend/commit/a4795b392c41f1a77bfa32a09df677630e23e495))
* **security:** overhaul Redis rate limiter middleware and introduce PoW engine ([82797e7](https://github.com/Doratou/Backend/commit/82797e7409e755c885580e64a74247c85b573551))
* **settlements:** add payment handlers and group status endpoint ([7416387](https://github.com/Doratou/Backend/commit/74163879b9640d312c3e73adcf0258562ae6003f))
* **settlements:** add PostgreSQL persistence and group lifecycle guards ([6f18b3b](https://github.com/Doratou/Backend/commit/6f18b3b13e34031bbc9449dbfc989f1d48ad9767))
* **settlements:** add settlement and group lifecycle migration ([4208035](https://github.com/Doratou/Backend/commit/420803553fc54831fe8712f8d9b4fd71c2088574))
* **settlements:** define domain models, enums and repository contract ([eb426ea](https://github.com/Doratou/Backend/commit/eb426eac193807845757477f8a230a3f97f08ede))
* **settlements:** implement payment workflows and receipt access rules ([5e36e1a](https://github.com/Doratou/Backend/commit/5e36e1adb437b35b6d89bfffee293c424a71b15a))
* **settlements:** merge payment workflows into develop ([222ce92](https://github.com/Doratou/Backend/commit/222ce92231fceb1e122c1e37cb415d7b965233fb))
* **settlements:** register payment and group status routes ([05707c1](https://github.com/Doratou/Backend/commit/05707c190ededd750665d6499bc69a5f74d6c1cb))
* **settlements:** wire payment dependencies in application startup ([8d0de89](https://github.com/Doratou/Backend/commit/8d0de89d07d7814d52eb7eb69952246c2cb264a5))
* **testing:** add comprehensive manual test seed data and optionize bruno requests ([af54dc8](https://github.com/Doratou/Backend/commit/af54dc81b78a954d03b53763c36aa914ce940acb))
* **user:** add admin account activation and reorganize dev tooling ([29d3558](https://github.com/Doratou/Backend/commit/29d35587dc0550e0bef47b3a65fefc67b93cf531))
* **user:** add RegenerateFriendCode in repository and service using database function ([1534d85](https://github.com/Doratou/Backend/commit/1534d853278f7ab4cdb7ebc0a003d5fb29650204))
* **user:** define DTOs and UserService adapter interface for user and admin operations ([f840a19](https://github.com/Doratou/Backend/commit/f840a197e41ecc3b0a6ab748ea4e11693bffb20c))
* **user:** expose POST /users/me/friend-code/regenerate endpoint and Bruno documentation ([a0fc956](https://github.com/Doratou/Backend/commit/a0fc956bab5d5505ee753fd950c372939517bf78))
* **user:** implement UserHandler endpoints for profile, contacts, wallet, and admin operations ([51d8ceb](https://github.com/Doratou/Backend/commit/51d8ceb794fce6969611f50034b5fc96590302dd))
* **user:** implement UserService methods for profile, contact changes, wallet, preview, and admin actions ([90fe862](https://github.com/Doratou/Backend/commit/90fe8627a59701da1fb5f27f43b281daacc6b70e))
* **users:** add friend codes and recipient lookup ([2796c85](https://github.com/Doratou/Backend/commit/2796c85f560638a88bfb1a85ac1cfbbcb9b74f3c))
* **users:** add persistent unique friend codes ([20c74e4](https://github.com/Doratou/Backend/commit/20c74e4599b03e033dcd6e5247ae32d310d3b0bc))
* **users:** add profile and friend code lookup endpoints ([dbe84eb](https://github.com/Doratou/Backend/commit/dbe84eb375b50113d38d59b4261fde58e1f799b9))
* **user:** update domain and postgres repository for user filters, role update and unscoped retrieval ([b80cbd1](https://github.com/Doratou/Backend/commit/b80cbd12c186d3a0cbdedab4d77082d0f795dc66))
* **user:** wire user profile, contact, wallet, preview, and admin routes in UserRouter ([a3c3393](https://github.com/Doratou/Backend/commit/a3c3393ae2b0d32bd600afc4c835686d95365cd4))
* **weather:** add weather service and dtos ([11c3907](https://github.com/Doratou/Backend/commit/11c3907032a411bab511db77a8f48e7526f0bff6))


### Bug Fixes

* **ai:** restore AvalAI configuration and format Go sources ([fceddd3](https://github.com/Doratou/Backend/commit/fceddd3a7f8185bc0f117c1e421c9fe02204c2b9))
* **ci:** fixed Trivy error for docs Dockerfile ([27eb867](https://github.com/Doratou/Backend/commit/27eb867c177fda515af0dd10997fb7771170ba90))
* **ci:** update vulnerable Go dependencies and format domain types ([5994aa8](https://github.com/Doratou/Backend/commit/5994aa8abe998bd3c898758585cc8994a8c5befc))
* **deps:** update go-jose to address JWK parsing vulnerabilities ([13766ac](https://github.com/Doratou/Backend/commit/13766ac784ea493fe7a6402c06e5d1c3e85d81eb))
* **migrations:** resolve version collisions and preserve existing schemas ([6d89505](https://github.com/Doratou/Backend/commit/6d89505221079eefb97666e37e31046ef329431d))
* **settlements:** allow rejecting payments without a reason ([7e77e33](https://github.com/Doratou/Backend/commit/7e77e333dfa3321074d93656fa4e6d54f319d204))
* **settlements:** remove cancellation reasons and align Bruno docs ([ff4ce80](https://github.com/Doratou/Backend/commit/ff4ce80c49a0a6e220415d5e14cd290c71521d9f))

## [0.1.0](https://github.com/Doratou/Backend/compare/v1.1.0...v0.1.0) (2026-11-09)


### Features

* **auth,http:** simplify OTP payload contract and add CORS middleware ([5ad5017](https://github.com/Doratou/Backend/commit/5ad501710dc7d1661ad4d0e6e7bb2fde56b1da70))
* **auth:** add passwordless otp login, password change and reset flows ([f1c45c1](https://github.com/Doratou/Backend/commit/f1c45c1d0bd90fdb58b1a1da9e4d1760e7f47d59))
* **auth:** implement otp and registration application services with contracts and env generator ([fe839ab](https://github.com/Doratou/Backend/commit/fe839ab16f700224440ac0ae283596456a15c7b0))
* **auth:** implement presentation layer and full auth lifecycle with login, logout, refresh, and blacklist middleware ([2ae5ad7](https://github.com/Doratou/Backend/commit/2ae5ad74d9ad02ca42cf13ed2af6f3833d8b0cf9))
* **auth:** send email OTPs via SMTP ([bd686ca](https://github.com/Doratou/Backend/commit/bd686ca26fa751c3018f59c0c97ceedecaa5a18c))
* **auth:** unify passwordless OTP login into a single endpoint ([1ac0598](https://github.com/Doratou/Backend/commit/1ac0598d8ae3da8740c79e2c0354f52c6af81cf2))
* **bruno:** add Bruno v4.2.1 API client collection for auth and health endpoints ([0b62033](https://github.com/Doratou/Backend/commit/0b62033b6be2b01e8bc2302779d190b2df38b51a))
* **captcha:** add captcha domain contract, config, errors, and base engine ([8583770](https://github.com/Doratou/Backend/commit/8583770e1c1d324e87fbc82ae566f5fb90e75b83))
* **captcha:** implement captcha service and redis repository with replay protection ([c89760c](https://github.com/Doratou/Backend/commit/c89760c96ecdaf819b270edc70ca9fab3c2b0ab5))
* **captcha:** implement presentation handler, guard middleware, and Bruno collection ([9d3eb33](https://github.com/Doratou/Backend/commit/9d3eb338b891ab89e9fc893c4ca330e6f69b3f3c))
* **captcha:** implement string captcha handling, guard middleware, and startup config logging ([7369c8a](https://github.com/Doratou/Backend/commit/7369c8a812edf688cc141e09f62f89f61e68c9f3))
* **cmd/app:** wire authentication, persistence, and presentation dependencies in bootstrap entrypoint ([f113ca9](https://github.com/Doratou/Backend/commit/f113ca92d68f8396ea7868358c7875769304df96))
* **infra:** implement postgres and redis persistence, transaction managers, jwt maker and password hasher ([c4b1a7f](https://github.com/Doratou/Backend/commit/c4b1a7f5ff852f950113307f2782329696c7247f))
* **migration:** add users and attachments schema with goose migration cli ([0383cc1](https://github.com/Doratou/Backend/commit/0383cc1954c676dda9271ce05045d53c5312b460))
* **recaptcha:** introduce Google reCAPTCHA v3 verification package ([942d41d](https://github.com/Doratou/Backend/commit/942d41da1b300a5d5b601ec0d7eff385ab7c5d61))
* **seed:** implement idempotent database seeder for default regular and admin users ([24fec69](https://github.com/Doratou/Backend/commit/24fec6904e5631c0560f1f6c635908450b44d875))


### Miscellaneous Chores

* reset release baseline ([234be1a](https://github.com/Doratou/Backend/commit/234be1ac0d45fb8931dffa9bd4f86e4c3cc4f54d))

## Changelog
