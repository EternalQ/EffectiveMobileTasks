# Ревью: Птуха Данила (G11) / Тема 10: Микросервисы
# Репозиторий: https://github.com/EternalQ/EffectiveMobileTasks.git (task10/)

## Общее впечатление

Все 7 заданий на месте, проект собран в единую архитектуру с docker-compose. WebSocket-хендлер падает с паникой в двух сценариях, плюс сетевой вызов RabbitMQ под мьютексом блокирует весь REST API. Ресурсы не закрываются в нескольких местах. Тема с ресурсами и корректным завершением повторяется из ревью message-brokers. Фундамент хороший, но нужно пофиксить.

## Что хорошо

- Грамотная структура проекта: `internal/service`, `pkg/rabbit`, `proto/`, `cmd/clients`. Видно понимание Go layout.
- JWT реализация: interceptor с bypass для Login, проверка signing method в Verify, токен через metadata. Параметризовано (secretKey, tokenDuration), а не захардкожено.
- Prometheus histogram для HTTP duration + middleware + `/metrics` endpoint. Ровно то, что нужно для мониторинга.
- docker-compose поднимает всю инфраструктуру: server, RabbitMQ, Prometheus, Grafana, ELK (elasticsearch + kibana + filebeat). Конфиги на месте.
- `sync.RWMutex` в REST-сервисе используется правильно: `RLock` для GetUsers, `Lock` для мутаций.
- Клиентское приложение с интерактивным меню для REST, gRPC, WebSocket и RabbitMQ. Удобно для демонстрации.

## Критичные замечания

### 1. WebSocket: нет return после ошибки Upgrade (`ws.go:18-22`)

Если `Upgrade` вернет ошибку (например, обычный HTTP-запрос на `/ws`), код продолжит выполнение с `ws == nil`. И `defer ws.Close()`, и `ws.ReadMessage()` упадут с nil pointer panic. Соединение hijacked, поэтому стандартный `recover()` из `net/http` может не спасти.

```go
// сейчас
ws, err := upgrader.Upgrade(w, r, nil)
if err != nil {
    log.Printf("error: %v", err)
}
defer ws.Close()

// надо
ws, err := upgrader.Upgrade(w, r, nil)
if err != nil {
    log.Printf("error: %v", err)
    return
}
defer ws.Close()
```

### 2. WebSocket: паника и ложный матчинг команды `/add` (`ws.go:33-36`)

Два бага в одном месте:

**Паника**: если клиент отправит `/add` без имени, `strings.Split("/add", " ")` вернет `["/add"]`, и `args[1]` упадет с index out of range.

**Ложный матчинг**: `strings.Contains(msg, "/add")` сработает на любом сообщении, где есть `/add` в любом месте. Например, `"download/additional"` или `"hello /add world"`. Во втором случае `args[1]` будет `"/add"`, а не имя.

```go
// сейчас
if strings.Contains(msg, "/add") {
    args := strings.Split(msg, " ")
    s.register(&models.User{Name: args[1]})

// надо: проверяем что команда в начале + есть аргумент
if strings.HasPrefix(msg, "/add ") {
    name := strings.TrimPrefix(msg, "/add ")
    if name == "" {
        ws.WriteMessage(websocket.TextMessage, []byte("usage: /add <name>"))
        continue
    }
    s.register(&models.User{Name: name})
    continue
}
```

## Важные замечания

### 3. Сетевой вызов RabbitMQ под мьютексом (`rest.go:101-112`)

`register()` захватывает `s.mu.Lock()`, а потом внутри лока вызывает `s.rClient.Pub()`. Это сетевой вызов. Если RabbitMQ тормозит или недоступен, все остальные HTTP-хендлеры (GetUsers, UpdateUser, DeleteUser) будут заблокированы, потому что ждут тот же мьютекс.

```go
// надо вынести Pub за пределы лока
func (s *RestService) register(user *models.User) {
    s.mu.Lock()
    s.lastId++
    user.ID = s.lastId
    s.users[user.ID] = *user
    s.mu.Unlock()

    if err := s.rClient.Pub("New user: " + user.Name); err != nil {
        s.log.Error("rabbit pub", "err", err)
    }
}
```

### 4. RabbitMQ subscriber: бесконечный цикл при закрытии канала (`cmd/main.go:58-62`)

```go
for {
    d := <-msgs
    fmt.Printf("[%v] Notif: %s\n", d.Timestamp, d.Body)
    time.Sleep(100 * time.Millisecond)
}
```

Если RabbitMQ отключится, канал `msgs` закроется. Receive из закрытого канала возвращает zero-value `Delivery` мгновенно и не блокируется. Цикл начнет крутиться с частотой 10 раз в секунду, печатая пустые уведомления и грея процессор. В ревью message-brokers была та же категория проблем (ресурсы, горутины, бесконечные циклы). Повторяющееся слабое место.

```go
for d := range msgs {
    fmt.Printf("[%v] Notif: %s\n", d.Timestamp, d.Body)
}
fmt.Println("Соединение закрыто")
```

### 5. RabbitClient.Close() закрывает connection до channel (`rabbit/client.go:43-46`)

```go
func (c *RabbitClient) Close() {
    c.conn.Close()  // соединение закрыто
    c.ch.Close()    // канал уже мертв, ошибка
}
```

Channel зависит от connection. Правильный порядок: сначала channel, потом connection.

```go
func (c *RabbitClient) Close() {
    c.ch.Close()
    c.conn.Close()
}
```

### 6. Ошибки json.Encode игнорируются в REST-хендлерах (`rest.go:98, 141, 175`)

В RegisterUser, UpdateUser и GetUsers результат `json.NewEncoder(w).Encode(...)` не проверяется. Это повторяющийся паттерн: в ревью go-testing и code-quality было то же самое. Если Encode упадет, клиент получит обрезанный JSON.

```go
if err := json.NewEncoder(w).Encode(user); err != nil {
    s.log.Error("encode response", "err", err)
}
```

### 7. http.ListenAndServe ошибка игнорируется (`main.go:60`)

Последняя строка `main()`:
```go
http.ListenAndServe(":"+restPort, rest)
```

Если порт занят или сервер упал, программа тихо завершится без ошибки.

```go
if err := http.ListenAndServe(":"+restPort, rest); err != nil {
    log.Fatal("rest server:", err)
}
```

### 8. gRPC connection не закрывается в клиенте (`cmd/clients/grpc.go:23-33`)

В `NewGrpcClient` создается `grpc.NewClient(addr, ...)`, но `conn` нигде не закрывается. Нужно сохранить conn в структуре и добавить метод Close.

```go
type GrpcClient struct {
    conn   *grpc.ClientConn
    client pb.UserServiceClient
    ctx    context.Context
}

func (c *GrpcClient) Close() {
    c.conn.Close()
}
```

## Рекомендации

### 9. WebSocket: нет поддержки бинарных сообщений

Задание просит "поддержку нескольких типов сообщений (например, текстовые и бинарные данные)". Сейчас все обрабатывается как текст. Можно добавить проверку типа:

```go
t, p, err := ws.ReadMessage()
// ...
switch t {
case websocket.TextMessage:
    // текущая логика
case websocket.BinaryMessage:
    log.Printf("Binary: %d bytes", len(p))
    ws.WriteMessage(t, p)
}
```

### 10. Notification-сервис не отдельный микросервис

По заданию 5 нужно "два микросервиса": user-service и notification-service. Сейчас notification это клиентский подписчик в `cmd/main.go` (case "3"), а не отдельный сервис. Для полноты решения стоило бы выделить notification в отдельный `cmd/notification/main.go` со своим Dockerfile.

### 11. Login: аутентификация только по имени, без пароля (`grpc.go:33-57`)

`LoginRequest` содержит только `name`, поля для пароля нет. Любой, кто знает имя пользователя, получает JWT-токен. Для учебного проекта это допустимо, но задание говорит "аутентификация", а auth по имени без пароля это фактически авторизация, не аутентификация. Добавить поле `password` в proto и хешировать через bcrypt было бы правильнее.

### 12. Админа можно удалить (`grpc.go:25, 91-100`)

Admin-пользователь с ID=0 создается при инициализации, но `DeleteUser` никак не защищает его. После удаления admin зайти в систему будет невозможно (Login не найдет ни одного пользователя с совпадающим именем).

### 13. Login берет write lock для чтения (`grpc.go:38`)

`Login` только читает `s.users` (ищет по имени), но использует `s.mu.Lock()` вместо `s.mu.RLock()`. Из-за этого Login блокирует все параллельные операции, хотя мог бы не мешать другим читателям.

### 14. REST и gRPC модели пользователя не согласованы

REST User: только `id` + `name`. gRPC User: `id` + `name` + `email`. Из-за этого пользователь, созданный через gRPC, имеет email, а через REST нет. Стоит привести к одной модели.

### 15. Go 1.22+ net/http: gorilla/mux можно заменить на stdlib

С Go 1.22 стандартный `http.ServeMux` поддерживает методы и path params:

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /users", s.RegisterUser)
mux.HandleFunc("GET /users", s.GetUsers)
mux.HandleFunc("PUT /users/{id}", s.UpdateUser)
mux.HandleFunc("DELETE /users/{id}", s.DeleteUser)
```

А вместо `mux.Vars(r)` используется `r.PathValue("id")`. У тебя в go.mod стоит Go 1.25, так что stdlib покрывает все нужды. Плюс в задании сказано "Go стандартные библиотеки для работы с HTTP".

### 16. `context.WithValue` со строковым ключом (`auth_grpc.go:82`)

```go
ctx = context.WithValue(ctx, "name", claims.Name)
```

Строковые ключи могут конфликтовать между пакетами. Принято использовать неэкспортируемый тип:

```go
type contextKey string
const nameKey contextKey = "name"
ctx = context.WithValue(ctx, nameKey, claims.Name)
```

## Итог

Зачтено с замечаниями (7/10)

Все 7 заданий представлены: REST CRUD, gRPC с протобафами, JWT auth, WebSocket, RabbitMQ коммуникация, Prometheus/Grafana/ELK, конфигурация через env. Архитектура правильная, docker-compose поднимает все. Два критичных места: паники в WebSocket-хендлере (nil после Upgrade, index out of range и ложный матчинг /add). Оба локализованы в одном хендлере и чинятся в пару строк. Из повторяющихся проблем: ресурсы не закрываются (gRPC conn, RabbitMQ в сервере), ошибки json.Encode игнорируются (третье ревью подряд), и бесконечный цикл при закрытии канала (та же тема, что в message-brokers). Обрати внимание на паттерн: корректное завершение ресурсов и обработка ошибок IO. После фиксов будет зачтено.
