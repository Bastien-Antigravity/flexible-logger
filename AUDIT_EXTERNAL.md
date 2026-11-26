
flexible-logger — repository audit
Repository: Bastien-Antigravity/flexible-logger · Branch: develop · Reviewed commit: cc391d3 · 24 November 2026

Strong architecture
 Audit reliability needs work

The library has a good overall structure: a focused logging engine, a common sink interface, fan-out support, asynchronous workers, multiple serializers and predefined profiles. The recent work on pooled-entry ownership, buffer-drop accounting and caller attribution is meaningful.

However, I would not yet treat the Audit profile as a dependable audit trail. Two implementation details are particularly important:

1. Remote write is not remote acknowledgement. NewAuditLogger describes its synchronous network sink as ensuring server receipt, but it writes serialized bytes to an io.Writer/TCP connection. A successful write does not prove that the log server received, validated or persisted the event. The public logger methods also return no error, so the application cannot fail a business operation when logging fails.
2. File rotation can destroy the active log on a rename failure. RollingFileWriter.rotate() ignores errors from renaming the active file, then opens that same path with O_TRUNC. If the rename fails but the file can still be opened, the existing log can be truncated. For an audit-oriented logger, this is a high-priority data-loss risk.

I’m checking how well the tests and CI actually verify these guarantees, and whether the recent reliability improvements close the other lifecycle and overload risks.
Findings from the deeper inspection

1. Critical — Audit mode can silently become non-audit mode
   The NewAuditLogger() constructor returns a logger even when its remote log server is unavailable or misconfigured. If both the primary and emergency file paths fail, it falls back to console output and still returns a usable logger.

There is no status or error returned to the calling application. In addition, the public logging methods and Close() return no errors, so the application cannot reliably detect that an audit record was not written.

This is a mismatch between the intended guarantee and the actual API contract.

Recommendation: provide a separate strict-audit constructor that returns an error unless required local and remote destinations are ready. Critical writes should expose delivery failures to the caller. Keep the existing best-effort behavior for ordinary application logging.

Source: audit.go · logger.go
2. High — Audit ordering is not guaranteed across destinations
MultiSink fans out each entry sequentially, but it has no global ordering mechanism across concurrent callers.

For example, two goroutines can submit records A and B concurrently. Each destination has its own synchronization, so the file sink and network sink can receive these records in different orders.

That matters when reconstructing an event sequence across multiple systems.

Recommendation: assign a monotonically increasing sequence number before dispatch, and introduce an ordering mechanism for strict audit streams. Test ordering across concurrent producers and across all destinations.

Source: multi.go
3. High — Async shutdown timeout does not prove worker termination
The AsyncSink.Close() implementation has improved substantially: it closes the queue and waits for the worker, with a configurable two-second default timeout.

However, when the timeout expires, it closes the downstream sink and returns. It does not subsequently establish that the worker has terminated.

The current timeout test illustrates the gap: its mock sink blocks indefinitely until the test itself releases it, after Close() has already returned. The test proves bounded return time, but not bounded goroutine lifetime or complete cleanup.

There is also a possible race in SetOnError(): the setter uses a mutex, but the worker reads OnError without taking that mutex.

Recommendation: make the worker lifecycle explicitly observable; ensure downstream cancellation actually interrupts pending writes; wait for worker termination within a documented overall deadline. Protect callback configuration consistently, preferably by configuring callbacks before starting the worker or using atomic publication.

Source: async.go · sink_test.go
4. Medium — Buffer-drop alerting can still cause a logging storm
The new DroppedCount() counter and throttled alert mechanism are good improvements. The mechanism limits repeated notification alerts.

However, every dropped log still returns an error from AsyncSink.Write(). LogEngine.logInternal() then calls ReportInternalError(), which writes synchronously to stderr.

Under sustained buffer saturation, this can produce one synchronous diagnostic per dropped message. The Telegram alert may be throttled, but the local error output is not.

This can amplify pressure precisely when the logging system is overloaded.

Recommendation: count and aggregate drop errors centrally. Emit a throttled diagnostic containing the number of dropped entries, rather than printing a diagnostic for every individual drop.

Source: async.go · log_engine.go
5. Medium — Standard profile can duplicate console output
In NewStandardLogger(), if opening the log file fails, fileSink is assigned to consoleSink.

The code subsequently appends both consoleSink and fileSink to MultiSink. They are then the same sink, meaning each log entry is written to the console twice.

This is a small but concrete fallback-path bug.

Source: standard.go
6. Medium — Text logs can be forged or misparsed
TextSerializer appends the message without escaping newline or carriage-return characters.

A message containing embedded line breaks can therefore produce output that looks like multiple log records. This can confuse log parsers or allow misleading entries to appear in text-based operational logs.

Recommendation: escape control characters in messages and metadata, or prefer structured JSON for logs that may contain untrusted input.

Source: text_serializer.go
7. Medium — Notification loss is not consistently observable
The remote notifier has a bounded queue of 1,000 messages. When full, Notify() returns an error, which is useful. But there is no corresponding drop counter and no throttled overflow accounting comparable to AsyncSink.

There is also a lifecycle edge case: if its connection worker exits after a failed connection initialization, the notifier can still accept messages into its channel even though no worker remains to send them.

The local notifier similarly returns an error on queue saturation but has no drop counter, despite its source comments describing one.

The operational contracts in AGENTS.md are therefore ahead of the implementations.

Sources: remote_notifier.go · local_notifier.go · AGENTS.md
8. Testing — Several important claims are not yet demonstrated
There is a meaningful test suite, and it is better than a project relying only on compilation. But I found some gaps between the documented coverage and the actual tests.
Area	What I found
All logger profiles	TestAllProfiles directly exercises only Minimal and Developer profiles
Audit blocking	Uses a synthetic SlowSink, not the actual audit network path
Audit fallback	The test's slash-containing name is sanitized into a filename; it does not reliably induce the intended file-creation failure
Remote delivery	Mock servers accept and discard TCP data; they do not verify remote persistence or acknowledgement
Shutdown	Timeout test checks return time, not worker termination
Performance	Benchmark counts submitted calls, not confirmed records received by the mock server

The emergency audit fallback also differs from the TODO description: the TODO specifies a fallback under os.TempDir(), while the implementation uses the original log directory's parent. That means the fallback can fail for the same underlying directory-permission problem.

Sources: profiles_test.go · sink_test.go · TODO.md · benchmark

What is particularly good
There are several design decisions I would preserve.

The ownership model is a strong point. LogEntry objects are pooled, with explicit Retain() and Release() operations. MultiSink and AsyncSink account for additional consumers instead of simply sharing pointers and hoping they remain valid. This is an important discipline for high-throughput Go code.

The sink abstraction is appropriate. The engine does not need to know whether it is writing to a file, console, TCP connection or buffered asynchronous sink. This makes it relatively easy to add new destinations without rewriting the logging engine.

The profiles make trade-offs visible. Standard, Audit, Cloud, HighPerf and Minimal are useful abstractions. Keeping synchronous and asynchronous policies separate is better than trying to make one configuration serve every workload.

The serialization architecture is improving. Moving Cap’n Proto schemas into microservice-toolbox is the right direction for maintaining wire compatibility between services.

The latest reliability work is substantive. The caller-attribution parity fix, observable drop counter, buffer-drop alert throttling and bounded shutdown wait are useful changes. These are more than cosmetic refactors.

The CI setup is reasonably strong. The repository delegates to the shared fleet workflow, which checks formatting, module integrity, builds and runs the Go tests with the race detector. The internal module dependencies are pinned to specific pseudo-versions rather than being fetched from a mutable branch.

One governance weakness remains: the reusable workflow is referenced as @develop, not by an immutable commit. Its behavior can therefore change independently of this repository.

Sources: entry.go · architecture overview · CI configuration · shared CI workflow · go.mod
Score
Overall engineering quality
7.2/10

My qualitative assessment of the inspected implementation, not a benchmark or external certification.

Architecture

8.8/10

Code organization

8.4/10

Memory ownership and pooling

8.0/10

Go implementation

7.8/10

Concurrency

7.3/10

Testing strategy

7.1/10

CI and build reproducibility

8.2/10

Performance potential

8.5/10

Performance evidence

6.3/10

Error handling

6.4/10

General reliability

6.5/10

Audit and durability guarantees

4.5/10

Production readiness

6.7/10

The important distinction is that the library's architecture and performance potential are stronger than its current reliability guarantees. The low audit/durability score is not a judgment that ordinary application logging is unusable; it reflects the gap between an ordinary logger and a system that can be relied on as an authoritative audit record.
Priority remediation plan
I would address the problems in this order.
Priority	Work	Expected outcome
P0	Fix rolling-file rotation and test rename/open failures	Prevent silent destruction of existing logs
P0	Define strict audit delivery semantics and fail-closed initialization	Make audit failures detectable by the application
P1	Add server receipt/persistence acknowledgements for audit records	Distinguish written, received and persisted
P1	Establish cross-sink ordering for audit streams	Preserve a reconstructable event sequence
P1	Correct AsyncSink shutdown and callback synchronization	Avoid lingering workers and races
P2	Aggregate overload diagnostics and add notifier drop counters	Prevent diagnostic amplification and silent notification loss
P2	Correct profile fallback duplication and test all profiles	Make fallback behavior predictable
P2	Add end-to-end serialization, delivery, durability and shutdown tests	Verify actual system guarantees
P3	Escape text log fields and improve dependency pinning	Improve log integrity and build governance

For audit-grade logging, I would also introduce an explicit durability policy. A synchronous Write() is not equivalent to durable storage: the file writer does not explicitly fsync/Sync() the file before reporting success. The policy should specify whether success means accepted by the process, written to the operating system, flushed to stable local storage, or acknowledged by the remote server.
Comparison with tele-remote
Using the same qualitative scoring approach as the previous audit:
Dimension	flexible-logger	tele-remote
Overall	7.2	7.8
Architecture	8.8	8.8
Concurrency	7.3	8.2
Reliability	6.5	7.4
Error handling	6.4	7.5
Production readiness	6.7	7.0

The two projects have different responsibilities, so this is not a direct performance comparison.

flexible-logger has a particularly good foundation in modularity, reusable interfaces and memory ownership. tele-remote currently has a stronger operational reliability story in the areas reviewed. The logger's weak point is not its overall architecture; it is the correctness of its delivery, failure and durability guarantees.
Final verdict
Good engineering foundation
I would classify flexible-logger as a promising, well-structured logging library suitable for continued integration and testing, but not yet as a fully production-hardened logging platform for critical audit records.

For ordinary service observability, the design is close to a solid production candidate, subject to testing in the target environment.

For audit trails, financial transaction evidence or other records where losing or reordering events is unacceptable, the current implementation is not sufficient.

My most important recommendation: do not redesign the library. Preserve the architecture and strengthen the contracts around delivery, ordering, durability, failure visibility and shutdown. Those are the areas where targeted engineering work will produce the biggest improvement.

One limitation of this review: I inspected the current source, tests, documentation and CI configuration, but did not execute go test, go test -race, benchmarks or fault-injection tests. The scores therefore reflect source-level assessment, not independently verified runtime behavior.
