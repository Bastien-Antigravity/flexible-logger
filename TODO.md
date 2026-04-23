- [ ] Implement auto-start logic for `log-server` and `notif-server` if the connection fails or needs a restart.
  - *Note*: `conn_manager` (via `microservice-toolbox`) now natively handles indefinite network retries with multiplicative backoff. The missing logic is strictly for launching the process itself on a local machine.
- [ ] Consider adding a unified program launcher that checks service location (via `ip_resolver`) and potentially communicates through `tele-remote`.
 and restart with a function parameter to choose how the reconneciton should behave (e.g. restart the process, try to reconnect, etc.)
- [ ] Refine `OnError` callback behavior for each logger profile:
  - Determine if we should trigger a full configuration refresh (`distributed-config`) on repeated connection failures.
  - Implement logic to potentially restart the destination server (`log-server`/`notif-server`) via `tele-remote` if reconnections keep failing.
  - Evaluate if certain profiles (e.g., `Audit`) should halt the application entirely if the error handler cannot recover the link after a specific threshold.
- [ ] Need to remove the microservice-toolbox import, to prevent problems, anyway it should be not used here