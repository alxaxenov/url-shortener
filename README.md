# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


## Результаты профилирования

### POST /api/shorten/batch
```azure
(base) aiaksenov@R4X0VXVM5J url-shortener % go tool pprof -top -diff_base=profiles/saveBatch/base.pprof profiles/saveBatch/result.pprof
File: ___5go_build_main_go
Type: inuse_space
Time: 2026-04-20 01:02:14 MSK
Showing nodes accounting for -484.02kB, 10.71% of 4518.18kB total
      flat  flat%   sum%        cum   cum%
  548.84kB 12.15% 12.15%   548.84kB 12.15%  compress/flate.(*compressor).initDeflate (inline)
 -525.43kB 11.63%  0.52%  -525.43kB 11.63%  github.com/jackc/pgx/v5/internal/stmtcache.NewLRUCache (inline)
  520.04kB 11.51% 12.03%   520.04kB 11.51%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).appendParam
 -514.38kB 11.38%  0.64%  -507.65kB 11.24%  github.com/alxaxenov/url-shortener/tree/v2/internal/service.(*ShortenerService).SaveBatch
 -513.31kB 11.36% 10.72%  -513.31kB 11.36%  github.com/jackc/pgx/v5/internal/stmtcache.StatementName
  512.23kB 11.34%  0.62%   512.23kB 11.34%  runtime.mallocgc
 -512.03kB 11.33% 10.71%  -512.03kB 11.33%  syscall.anyToSockaddr
  512.02kB 11.33%  0.62%   512.02kB 11.33%  internal/sync.newEntryNode[go.shape.interface {},go.shape.interface {}] (inline)
 -512.02kB 11.33% 10.71%  -512.02kB 11.33%  sync.(*Pool).pinSlow
         0     0% 10.71%   548.84kB 12.15%  compress/flate.(*compressor).init
         0     0% 10.71%   548.84kB 12.15%  compress/flate.NewWriter (inline)
         0     0% 10.71%   548.84kB 12.15%  compress/gzip.(*Writer).Write
         0     0% 10.71%     6.73kB  0.15%  database/sql.(*DB).ExecContext
         0     0% 10.71%     6.73kB  0.15%  database/sql.(*DB).ExecContext.func1
         0     0% 10.71%  -525.43kB 11.63%  database/sql.(*DB).PingContext
         0     0% 10.71%  -525.43kB 11.63%  database/sql.(*DB).PingContext.func1
         0     0% 10.71%  -525.43kB 11.63%  database/sql.(*DB).conn
         0     0% 10.71%     6.73kB  0.15%  database/sql.(*DB).exec
         0     0% 10.71%     6.73kB  0.15%  database/sql.(*DB).execDC
         0     0% 10.71%     6.73kB  0.15%  database/sql.(*DB).execDC.func2
         0     0% 10.71%  -518.70kB 11.48%  database/sql.(*DB).retry
         0     0% 10.71%     6.73kB  0.15%  database/sql.ctxDriverExec
         0     0% 10.71%     6.73kB  0.15%  database/sql.withLock
         0     0% 10.71%   512.02kB 11.33%  encoding/json.(*encodeState).marshal
         0     0% 10.71%   512.02kB 11.33%  encoding/json.(*encodeState).reflectValue
         0     0% 10.71%   512.02kB 11.33%  encoding/json.Marshal
         0     0% 10.71%   512.02kB 11.33%  encoding/json.typeEncoder
         0     0% 10.71%   512.02kB 11.33%  encoding/json.valueEncoder
         0     0% 10.71%  -525.43kB 11.63%  github.com/alxaxenov/url-shortener/tree/v2/internal/config/db/pg.(*ConnectorPG).Open
         0     0% 10.71%     4.37kB 0.097%  github.com/alxaxenov/url-shortener/tree/v2/internal/handler.(*ShortenerHandler).SaveBatch
         0     0% 10.71%  -512.03kB 11.33%  github.com/alxaxenov/url-shortener/tree/v2/internal/handler.Serve
         0     0% 10.71%   548.84kB 12.15%  github.com/alxaxenov/url-shortener/tree/v2/internal/middleware.(*UserMiddleware).Use.func1
         0     0% 10.71%   548.84kB 12.15%  github.com/alxaxenov/url-shortener/tree/v2/internal/middleware.(*compressWriter).Write
         0     0% 10.71%   548.84kB 12.15%  github.com/alxaxenov/url-shortener/tree/v2/internal/middleware.(*loggingResponseWriter).Write
         0     0% 10.71%   548.84kB 12.15%  github.com/alxaxenov/url-shortener/tree/v2/internal/middleware.GzipMiddleware.func1
         0     0% 10.71%   548.84kB 12.15%  github.com/alxaxenov/url-shortener/tree/v2/internal/middleware.WithLogging.func1
         0     0% 10.71%     6.73kB  0.15%  github.com/alxaxenov/url-shortener/tree/v2/internal/repository/db.(*DBRepo).SaveBatch
         0     0% 10.71%   548.84kB 12.15%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 10.71%   548.84kB 12.15%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 10.71%   548.84kB 12.15%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 10.71%     6.73kB  0.15%  github.com/jackc/pgx/v5.(*Conn).Exec
         0     0% 10.71%     6.73kB  0.15%  github.com/jackc/pgx/v5.(*Conn).exec
         0     0% 10.71%   520.04kB 11.51%  github.com/jackc/pgx/v5.(*Conn).execPrepared
         0     0% 10.71%   520.04kB 11.51%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).Build
         0     0% 10.71%  -525.43kB 11.63%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 10.71%  -525.43kB 11.63%  github.com/jackc/pgx/v5.connect
         0     0% 10.71%     6.73kB  0.15%  github.com/jackc/pgx/v5/stdlib.(*Conn).ExecContext
         0     0% 10.71%  -525.43kB 11.63%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
         0     0% 10.71%  -512.03kB 11.33%  internal/poll.(*FD).Accept
         0     0% 10.71%  -512.03kB 11.33%  internal/poll.accept
         0     0% 10.71%   512.02kB 11.33%  internal/sync.(*HashTrieMap[go.shape.interface {},go.shape.interface {}]).LoadOrStore
         0     0% 10.71%  -512.02kB 11.33%  log.(*Logger).output
         0     0% 10.71%  -512.02kB 11.33%  log.Println (inline)
         0     0% 10.71%  -512.02kB 11.33%  log.getBuffer (inline)
         0     0% 10.71% -1037.46kB 22.96%  main.main
         0     0% 10.71% -1037.46kB 22.96%  main.run
         0     0% 10.71%  -512.02kB 11.33%  main.run.func1
         0     0% 10.71%  -512.03kB 11.33%  net.(*TCPListener).Accept
         0     0% 10.71%  -512.03kB 11.33%  net.(*TCPListener).accept
         0     0% 10.71%  -512.03kB 11.33%  net.(*netFD).accept
         0     0% 10.71%  -512.03kB 11.33%  net/http.(*Server).ListenAndServe
         0     0% 10.71%  -512.03kB 11.33%  net/http.(*Server).Serve
         0     0% 10.71%   548.84kB 12.15%  net/http.(*conn).serve
         0     0% 10.71%   548.84kB 12.15%  net/http.(*timeoutHandler).ServeHTTP
         0     0% 10.71%     4.37kB 0.097%  net/http.(*timeoutHandler).ServeHTTP.func1
         0     0% 10.71%   553.22kB 12.24%  net/http.HandlerFunc.ServeHTTP
         0     0% 10.71%  -512.03kB 11.33%  net/http.ListenAndServe (inline)
         0     0% 10.71%   548.84kB 12.15%  net/http.serverHandler.ServeHTTP
         0     0% 10.71% -1037.46kB 22.96%  runtime.main
         0     0% 10.71%   512.23kB 11.34%  runtime.malg
         0     0% 10.71%   512.23kB 11.34%  runtime.newobject
         0     0% 10.71%   512.23kB 11.34%  runtime.newproc.func1
         0     0% 10.71%   512.23kB 11.34%  runtime.newproc1
         0     0% 10.71%   512.23kB 11.34%  runtime.systemstack
         0     0% 10.71%   512.02kB 11.33%  sync.(*Map).LoadOrStore (inline)
         0     0% 10.71%  -512.02kB 11.33%  sync.(*Pool).Get
         0     0% 10.71%  -512.02kB 11.33%  sync.(*Pool).pin
         0     0% 10.71%  -512.03kB 11.33%  syscall.Accept
```

### GET /api/user/urls
```azure
(base) aiaksenov@R4X0VXVM5J url-shortener % go tool pprof -top -diff_base=profiles/userURLs/base.pprof profiles/userURLs/result.pprof
File: ___5go_build_main_go
Type: inuse_space
Time: 2026-04-20 02:48:56 MSK
Showing nodes accounting for -1.02MB, 26.15% of 3.90MB total
      flat  flat%   sum%        cum   cum%
   -0.51MB 13.14% 13.14%    -0.51MB 13.14%  github.com/jackc/pgx/v5/internal/stmtcache.NewLRUCache (inline)
   -0.51MB 13.01% 26.15%    -0.51MB 13.01%  runtime.mallocgc
   -0.50MB 12.81% 38.96%    -0.50MB 12.81%  context.(*cancelCtx).Done
    0.50MB 12.81% 26.15%     0.50MB 12.81%  time.map.init.0
         0     0% 26.15%    -0.51MB 13.14%  database/sql.(*DB).PingContext
         0     0% 26.15%    -0.51MB 13.14%  database/sql.(*DB).PingContext.func1
         0     0% 26.15%    -0.51MB 13.14%  database/sql.(*DB).conn
         0     0% 26.15%    -0.50MB 12.81%  database/sql.(*DB).connectionOpener
         0     0% 26.15%    -0.51MB 13.14%  database/sql.(*DB).retry
         0     0% 26.15%    -0.51MB 13.14%  github.com/alxaxenov/url-shortener/tree/v2/internal/config/db/pg.(*ConnectorPG).Open
         0     0% 26.15%    -0.51MB 13.14%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 26.15%    -0.51MB 13.14%  github.com/jackc/pgx/v5.connect
         0     0% 26.15%    -0.51MB 13.14%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
         0     0% 26.15%    -0.51MB 13.14%  main.main
         0     0% 26.15%    -0.51MB 13.14%  main.run
         0     0% 26.15%     0.50MB 12.81%  runtime.(*scavengerState).init
         0     0% 26.15%    -0.50MB 12.83%  runtime.allocm
         0     0% 26.15%     0.50MB 12.81%  runtime.bgscavenge
         0     0% 26.15%     0.50MB 12.81%  runtime.doInit (inline)
         0     0% 26.15%     0.50MB 12.81%  runtime.doInit1
         0     0% 26.15%     0.50MB 12.83%  runtime.goexit0
         0     0% 26.15%    -0.01MB  0.34%  runtime.main
         0     0% 26.15%    -0.50MB 12.83%  runtime.mcall
         0     0% 26.15%    -0.50MB 12.83%  runtime.newm
         0     0% 26.15%    -0.51MB 13.01%  runtime.newobject
         0     0% 26.15%       -1MB 25.66%  runtime.park_m
         0     0% 26.15%    -0.51MB 12.98%  runtime.procresize
         0     0% 26.15%       -1MB 25.66%  runtime.resetspinning
         0     0% 26.15%    -0.51MB 12.98%  runtime.rt0_go
         0     0% 26.15%    -0.51MB 12.98%  runtime.schedinit
         0     0% 26.15%    -0.50MB 12.83%  runtime.schedule
         0     0% 26.15%    -0.50MB 12.83%  runtime.startm
         0     0% 26.15%    -0.50MB 12.83%  runtime.wakep
         0     0% 26.15%     0.50MB 12.81%  time.init
```
