Logrus Hooks
============

[Logrus](https://github.com/sirupsen/logrus) is a popular structured logger that allows to add hooks that are invoked for every message. These hooks can be used to send the messages to remote tracking and reporting system. Things may go wrong when talking to remote systems, like delays and errors.

The hooks from this package are [decorators](https://en.wikipedia.org/wiki/Decorator_pattern) that help to deal with:

* Transient errors
* Excessive logging messages
* Slow transmission times

### Setup

The examples will use the [syslog hook](https://github.com/sirupsen/logrus/tree/master/hooks/syslog) to send messages to _syslog_ daemon. The setup looks like this:

```go
import (
  "log/syslog"

  "github.com/misho-kr/logrus-hooks"
  "github.com/sirupsen/logrus"
  lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

func init() {

	// create a hook to be added to an instance of logger
	hook, err := NewSyslogHook("udp", "server.fqdn:514", syslog.LOG_DEBUG, "")
}
```

From here additional hooks can be added for enhanced logging functionality.

### Retry with backoff

Prevent transient errors from procesing the log messages with _retries_

```go
log.AddHook(RetryHook(
	hook,
	100 * time.Millisecond,  // delay between retries
	hooks.Retries(3),        // retry 3 times
	hooks.FactorPct(100),    // increase delay by 100% between retries
	hooks.JitterPct(10),     // jitter of 10% to the delay between retries
))
```

### Rate limits

Put a cap on the number of logging messages per time interval

```go
log.AddHook(RateLimitHook(
	hook,
	hooks.PerSecond(10),     // 10 messages per second
	hooks.Burst(20),         // burst of 20 messages allowed
))
```

### Asynchronous execution

Fire the hook in separate goroutine to avoid blocking the logger and main application

```go
log.AddHook(AsyncHook(
	hook,
	hooks.Senders(10),       // 10 goroutines to send messages
	hooks.BoostSenders(20),  // up to 20 additional goroutines when needed 
))
```
