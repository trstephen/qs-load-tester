Quoteserver Load Tester
====

How does the legacy quoteserver respond when under load? Let's find out!

This tool will make a new, concurrent request to the quoteserver for a random stock at a chosen interval. Responses can be written to a file for further analysis.

### Usage
**stdout** gets the response from the quoteserver. **stderr** will get stats at the end of the run (and errors). Redirect as you please~

```sh
# Make a new request every 20ms, sustain for 60s.
# Connect to quoteserve.seng.uvic.ca:4440 by default.
qs-load-tester --delay=20 --length=600 1> quotes.txt
```

Stats include
- *Quotes*: Number of quotes returned
- *Requests*: Number of requests emitted. Will be >= *Quotes*
- *Req per sec*: Estimate of request volume.

Use `qs-load-tester --help` for a wonderful list of capabilities.
