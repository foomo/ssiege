# About

Session siege **ssiege** is a http benchmarking and load testing utility with session support. It tries to fill the gap between [ab](//httpd.apache.org/docs/current/programs/ab.html),  [siege](//www.joedog.org/siege-home/) and [jmeter](//jmeter.apache.org jmeter).

We wrote it, because ab and siege do not support sessions and jmeter is too complicated.

It is the right tool if you are trying to benchmark a stateful web application server.

# Usage

You model user sessions in json config files and run them with the following command:

```bash
ssiege path/to/session-a.conf path/to/session-b.conf ... path/to/session-x.conf
```

While the benchmark is running it will print information to std out. You can also watch the running test in a web interface that is available on http://127.0.0.1:9999/ .

The benchmark will run until you send an interrupt signal to it, that would be a ctrl-c in most cases.

# Session config file

```json
{
    "concurrency" : 10,
    "name"        : "an example session",
    "color"       : "#ff0000",
    "session": {
        "server": "https://example.com",
        "calls" : [
            {
                "method"   : "GET",
                "uri"      : "/",
                "comment"  : "home page visit"
            },
            {
                "method"   : "POST",
                "mimetype" : "application/json",
                "uri"      : "/path/to/a/rest/service",
                "data"     : "{\"foo\":\"bar\"}",
                "comment"  : "this is how you perform XHR requests"
            },
            {
                "method"   : "GET",
                "uri"      : "/another/path",
                "comment"  : "look we have debugging support and we are going to wait for 200 ms",
                "wait"     : 200,
                "debug"    : true
            }
        ]
    }
}
```

# Web frontend

While running **ssiege** will start a web interface, that lets you see the current state of the running benchmark.
