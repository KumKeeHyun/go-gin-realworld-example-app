# go-gin-realworld-example-app

Example real world backend API built with Golang + Gin + Gorm + Swagger

## 목표

- gin, gorm 리마인드
- Hexagonal Architectural 공부
- DI 라이브러리 wire 사용
- Test Mock 라이브러리 gomock 사용
- Observability 툴 사용
- CI/CD 구축

## TODO

- service mock test
- prometheus, jaeger
- ci/cd

### 참고자료

- Hexagonal Architectural
  - https://medium.com/@matiasvarela/hexagonal-architecture-in-go-cfd4e436faa3
- 멱등성 
  - https://yozm.wishket.com/magazine/detail/2106/
  - 어댑터 쪽에서 구현할 예정
- transaction
  - https://github.com/dipeshhkc/Golang-Gorm-MultiLayer-DB-Transaction/tree/main