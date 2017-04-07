# gopools
Offshooot of github.com/fatih/pool hacked to support sessions and made slightly more configurable, additional capabilities are:
  - Session management, provides additional abstractions like Sessions to manage connections better, some of features of sessions are:
  - Duplicate recognition sessions
  - Commands run history
  - Capability to flush session commands run outside
  - Pre and Post configurators
  - Configurable from outside connections.
  - Adjustments to how factory is invoked for each connection to be live.
  - Pre and Post initializers if any available to run for each of the resources
  - Connection `PoolConn`
  - `Pool`
  - Ping and Keep-Alive semantics added
  - Connection and Pool Debug Mode enables more logs to show incase a leak is suspected
from the client application.

# To grab for use -
```go
go get -t github.com/AnirudhVyas/goopools

```

# Build instructions
```go
go build -v github.com/AnirudhVyas/gopools/...
```

NOTE: This needs more testing, additional suggestions, pull requests welcome.
