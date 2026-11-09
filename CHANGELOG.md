# Changelog

## [1.1.0](https://github.com/Doratou/Backend/compare/v1.0.0...v1.1.0) (2026-11-09)


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
* **domain:** define user and attachment entities, enums, errors and repository contracts ([193ca43](https://github.com/Doratou/Backend/commit/193ca43ec4bda2e69b16592997a6a236aa254de2))
* **infra:** implement postgres and redis persistence, transaction managers, jwt maker and password hasher ([c4b1a7f](https://github.com/Doratou/Backend/commit/c4b1a7f5ff852f950113307f2782329696c7247f))
* **migration:** add users and attachments schema with goose migration cli ([0383cc1](https://github.com/Doratou/Backend/commit/0383cc1954c676dda9271ce05045d53c5312b460))
* **recaptcha:** introduce Google reCAPTCHA v3 verification package ([942d41d](https://github.com/Doratou/Backend/commit/942d41da1b300a5d5b601ec0d7eff385ab7c5d61))
* **seed:** implement idempotent database seeder for default regular and admin users ([24fec69](https://github.com/Doratou/Backend/commit/24fec6904e5631c0560f1f6c635908450b44d875))

## Changelog
