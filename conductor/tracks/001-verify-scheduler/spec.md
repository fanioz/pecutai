# Spec: Verify and enhance Go scheduler cron evaluations

## 1. Problem Statement
We recently purged the dead `nex-insights` and `nex-notifications` cron specifications from the codebase. We need to verify that:
* The scheduler handles the remaining registered crons correctly.
* No reference to the deleted crons remains in the active database, system tests, or evaluations.
* The evaluation script `scripts/cron_eval/eval.py` functions correctly and reports clean status.

## 2. Requirements
* All Go tests under `internal/team/` must pass cleanly.
* Execution of `scripts/cron_eval/eval.py` must succeed with zero failures.
* Inspect cron configurations in `internal/team/broker_scheduler.go` to ensure intervals are correct.
