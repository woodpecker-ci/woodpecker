# Agent configuration

Agent configuration has the following command line variables that can be overridden via environmental variables

- "WOODPECKER_SERVER",
  - Usage:   "server address",
  - Value:   "localhost:9000",
- "WOODPECKER_USERNAME"
  - Usage:   "auth username",
  - Value:   "x-oauth-basic",
- "WOODPECKER_AGENT_SECRET"
  - Usage:   "server-agent shared password",        Value:   ""
- "WOODPECKER_DEBUG"
  - Usage:   "enable agent debug mode",
  - Value:   true,
- "WOODPECKER_LOG_LEVEL"
  - Usage:   "set logging level",
        "WOODPECKER_DEBUG_PRETTY"
  - Usage:   "enable pretty-printed debug output",
- "WOODPECKER_DEBUG_NOCOLOR"
  - Usage:   "disable colored debug output",
  - Value:   true,
- "WOODPECKER_HOSTNAME"
  - Usage:   "agent hostname",
        "WOODPECKER_PLATFORM"
  - Usage:   "restrict builds by platform conditions",
  - Value:   "linux/amd64",
- "WOODPECKER_FILTER"
  - Usage:   "filter expression to restrict builds by label",
- "WOODPECKER_MAX_PROCS"
  - Usage:   "agent parallel builds",
  - Value:   1,
- "WOODPECKER_HEALTHCHECK"
  - Usage:   "enable healthcheck endpoint",
  - Value:   true,
- "WOODPECKER_KEEPALIVE_TIME"
  - Usage:   "after a duration of this time of no activity, the agent pings the server to check if the transport is still alive",
- "WOODPECKER_KEEPALIVE_TIMEOUT"
  - Usage:   "after pinging for a keepalive check, the agent waits for a duration of this time before closing the connection if no activity",
  - Value:   time.Second * 20,
- "WOODPECKER_GRPC_SECURE"
  - Usage:   "should the connection to WOODPECKER_SERVER be made using a secure transport", 	
- "WOODPECKER_GRPC_VERIFY"
  - Usage:   "should the grpc server certificate be verified, only valid when WOODPECKER_GRPC_SECURE is true",
  - Value:   true,
