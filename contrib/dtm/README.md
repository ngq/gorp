# contrib/dtm

## Positioning

- This directory carries gorp distributed transaction backends.
- Current goal is not to rebuild a full DTM product layer inside gorp.
- Current goal is: provide a stable standard transaction integration layer, keep the default path lightweight and usable, and let advanced users go directly to official DTM SDK or their own DTM platform.

## Current Status

- `Partially available`: `dtmsdk`

## Explanation

- `dtmsdk` currently already has a minimum usable loop for `SAGA submit / query`.
- `TCC / XA / Barrier` are available as minimum bridge-level adapters with validation, submit path and basic behavior tests, but they are still not a full distributed transaction product layer.
- `framework/provider/dtm/noop` is a kernel default fallback capability and does not belong to unfinished contrib here.

## Native Escape Hatch

- `dtmsdk` is intentionally kept as a lightweight HTTP adapter, not a full mirror of official DTM SDK.
- The default adapter now provides unified down-dive conventions:
  1. `As(target any) bool`
  2. `Underlying() any`
- It also exposes `HTTPClient()` for users who need to customize transport-level behavior.
- For advanced orchestration semantics, richer barrier behaviors, or full official SDK features, users should directly use the official DTM SDK on top of or beside this adapter.

## Current Boundary

- `SAGA` is the current clearly supported main path.
- `TCC / XA / Barrier` are supported as standard transaction integration entry points, but not promoted as full official SDK replacement.
- `Query` is supported for transaction status inspection.
- HTTP layer already covers:
  - transient error retry
  - permanent error no-retry
  - context cancel stops retry

## Current P2 Closure

- `SAGA` remains minimum usable main path with submit, query and branch option tests.
- `TCC / XA` have been lifted from pure skeletons to minimum verifiable bridge state: builder validation, submit path and fake server verification all exist.
- `Barrier` has been lifted to minimum identity validation plus clear callback semantics, and exposes `BarrierContext` to business callback.
- `dtmsdk` now also provides unified escape hatch and transport access for advanced users.

## Production Claim

- This package is suitable as a stable standard integration layer when you want gorp to:
  - allocate gid
  - submit standard DTM transactions
  - query transaction status
  - keep transaction entrypoints unified through gorp contract
- This package does not claim to replace the full official DTM SDK.
- If a business needs deeper timeout/deadlock recovery policies, richer barrier semantics, or custom orchestration rules, that logic should be implemented by business code or by directly using official DTM SDK.

## P0 Constraints

- Do not document `dtmsdk` as “full DTM capability”.
- Always distinguish `SAGA` main path from advanced transaction modes.
- Keep public contract stable, and keep advanced capability outside the default adapter unless there is a clear production need.
