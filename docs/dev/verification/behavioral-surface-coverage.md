# Behavioral Surface Coverage Contract

## Purpose

RecompHamr requires **100% behavioral surface coverage** in addition to 100% Go statement coverage. Executing every statement is not proof that every contract, state transition, failure mode, or compatibility behavior has been tested.

This contract applies equally to:

- retained upstream behavior;
- modified behavior;
- replacements and rewrites;
- RecompHamr-Legacy parity work;
- optimizations or refactors with observable consequences;
- newly added behavior.

No old surface is grandfathered and no new surface is exempt.

## What is a behavioral surface?

A behavioral surface is any observable or relied-on contract, including as applicable:

- commands, arguments, options, exit codes, and help;
- TUI inputs, states, transitions, rendering contracts, cancellation, resize, and terminal restoration;
- tools, schema fields, outputs, errors, permissions, timeouts, and cleanup;
- configuration keys, defaults, validation, precedence, environment variables, and migration behavior;
- files, history, logs, caches, persistence, and recovery;
- provider and model transport behavior;
- agent-loop states and tool-dispatch behavior;
- public or cross-package APIs;
- platform-specific behavior;
- protocols, MCP, skills, discovery, precedence, activation, lifecycle, and trust boundaries when those stages are implemented;
- security boundaries and failure behavior;
- compatibility promises inherited from upstream or Legacy.

## Required inventory fields

Each active work packet or parity record must identify every in-scope surface and map it to:

- **Surface ID** — stable identifier.
- **Contract** — behavior that must be true.
- **Origin** — retained upstream, modified, replacement, Legacy parity, or new.
- **Owner** — package/service responsible for the behavior.
- **Tests** — exact tests proving the contract.
- **Applicable categories** — success, failure, malformed input, boundary, cancellation, timeout, cleanup, platform, Unicode, persistence, migration, compatibility, concurrency, and security as applicable.
- **Documentation** — exact user/developer/help/API documentation.
- **Runtime/manual evidence** — when automation cannot prove the contract.
- **Status** — complete, blocked, unsupported, or unverified.

A test category may be marked not applicable only with a concrete recorded reason.

## Closure rule

A task, phase, or parity row cannot close until every in-scope surface is complete. A passing statement-coverage profile, build, screenshot, or happy-path test cannot substitute for missing behavioral-surface evidence.

## Relationship to documentation coverage

Every surface in the inventory must also satisfy **100% meaningful documentation coverage**. Every Go package and exported symbol requires appropriate Go documentation, and every relied-on user, integration, configuration, persistence, security, lifecycle, and extension contract must be documented. Trivial private locals do not require artificial comments.
