## Исправления

- опечатки
- `WriteHeader` после ошибки `Encode()`

## Добавлено

- кейсы в `TestSubscription_Parse()` для невалидных дат
- тест и бенч для `Format()`
- тест для `CalculatePrice()` и необходимый интерфейс репозитория
```
go test ./... -cover
        github.com/EternalQ/effective-mobile-test               coverage: 0.0% of statements
        github.com/EternalQ/effective-mobile-test/docs          coverage: 0.0% of statements
        github.com/EternalQ/effective-mobile-test/pkg/api               coverage: 0.0% of statements
        github.com/EternalQ/effective-mobile-test/pkg/db                coverage: 0.0% of statements
ok      github.com/EternalQ/effective-mobile-test/pkg/models    (cached)        coverage: 100.0% of statements
ok      github.com/EternalQ/effective-mobile-test/pkg/service   0.003s  coverage: 64.3% of statements

go test ./... -bench .
?       github.com/EternalQ/effective-mobile-test       [no test files]
?       github.com/EternalQ/effective-mobile-test/docs  [no test files]
?       github.com/EternalQ/effective-mobile-test/pkg/api       [no test files]
?       github.com/EternalQ/effective-mobile-test/pkg/db        [no test files]
goos: linux
goarch: amd64
pkg: github.com/EternalQ/effective-mobile-test/pkg/models
cpu: 12th Gen Intel(R) Core(TM) i5-12450HX
BenchmarkSubscription_Format-12         13983897                81.58 ns/op
PASS
ok      github.com/EternalQ/effective-mobile-test/pkg/models    1.145s
time=2026-03-06T16:35:23.557+03:00 level=ERROR msg="Error while calculating price" where=service/SubscriptionService source=db/SubcriptionRepo.List method=CalculatePrice
PASS
ok      github.com/EternalQ/effective-mobile-test/pkg/service   0.004s
```