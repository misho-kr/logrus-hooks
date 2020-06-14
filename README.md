Logrus Hooks
============

[Logrus](https://github.com/sirupsen/logrus) is a popular structured logger. One of its many useful features is the option to add hooks that are invoked for every logging message.

Hooks can be used to send logging message to remote tracking and reporting system. Things can go wrong when talking to remote systems, like delays and errors. The hooks from this module provide means to deal with:

* Transient errors
* Excessive logging messages
* Slow transmission times

### Setup

To send messages to _syslog_ the setup is easy:

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
```