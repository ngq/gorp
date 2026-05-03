# contrib/configsource

## Native Escape Hatch

- `configsource` does not require every provider to mechanically expose native capability. Only real third-party client/SDK implementations that users might need vendor-specific features from should provide a unified escape hatch.
- `apollo / polaris / nacos / consul / etcd / kubernetes` such real external adapters, upon entering default delivery, should all have unified native escape hatch.
- `local / noop / fake` such kernel default capabilities or test stubs do not need forced escape hatch methods.
- The goal of escape hatch is not to mirror the full SDK surface into facade, but to guarantee:
  1. The default unified interface stays lightweight.
  2. Advanced users can get the underlying native client/SDK when needed.
  3. Vendor-specific capabilities do not leak into the public contract.
- Recommended unified convention priority:
  1. `As(target any) bool`
  2. `Underlying() any`
- Not recommended for each provider to invent completely different public escape method names.

## Official SDK First

- Starting from the current stage, `configsource` default direction: prefer official SDK / official Go client to carry complex semantics, rather than long-term self-built equivalent HTTP clients.
- First tier: `nacos`, `apollo`, `polaris`
- Second tier: `kubernetes`
- Blocking note: `apollo / polaris / kubernetes` have already switched to official SDK / official Go client, but still retain "partially available" claim until more complete publish/push/governance semantics are completed.

This directory carries gorp's config source backend implementations, but different subdirectories have inconsistent completion levels and must be distinguished by real status.

## Current Status

- `Done`: `consul`, `etcd`
- `Partially available`: `apollo`, `nacos`, `kubernetes`, `polaris`

## Explanation

- `Done` means having real config source access or minimum usable loop.
- `Partially available` means having completed the current stage minimum loop, with real `Load / Watch` main flow and behavior tests, but not yet to full productization.
- `Placeholder not done` means directory and provider wiring exist, but key main flow like Load / Watch is still TODO, empty implementation, or not actually done.
- `framework/provider/configsource/local` and `framework/provider/configsource/noop` are kernel default capabilities, not "placeholder not done contrib".

## Current Stage Boundary

- `nacos`: Already switched to `nacos-sdk-go/v2` for default `Load / Watch / Set` main flow, and retains fake client injection, initial callback, not found, publish failure, close-refuse-watch, and native escape hatch.
- `kubernetes`: Default implementation already switched to `client-go` `ConfigMaps Get/Watch`, and retains fake client injection, not found, source error, set-not-supported, close-refuse-watch, and native escape hatch.
- `apollo`: Default implementation already switched to official `agollo/v4`, and retains fake client injection and native escape hatch; current stage already completes initial load, change watch, duplicate revision suppression, error classification and retry/stop-retry boundary; but still claims "partially available".
- `polaris`: Default implementation already switched to official `polaris-go` `ConfigAPI`, and retains fake client injection and native escape hatch; current stage already completes initial load, change watch, duplicate revision suppression, error classification, retry/stop-retry boundary, and poll fallback; but still claims "partially available".
- All four items currently only claim "partially available" externally, and do not prematurely promote as complete config center productization.

## Current P2 Progress

- `apollo`: On the basis of P2 second-layer reinforcement, the default implementation has been further switched to official `agollo/v4`, with official SDK carrying default loading and change callbacks; simultaneously retains fake client injection, initial load, watch retry, duplicate revision suppression, AuthFailed/ConfigNotFound/SourceUnavailable error classification, poll fallback, native As/Underlying and behavior tests.
- `polaris`: On the basis of P2 second-layer reinforcement, the default implementation has been further switched to official `polaris-go` `ConfigAPI`, with official SDK carrying default loading and config change watch; simultaneously retains fake client injection, initial load, watch retry, duplicate revision suppression, AuthFailed/ConfigNotFound/SourceUnavailable error classification, poll fallback, native As/Underlying and behavior tests.
- `kubernetes`: Already completed second-tier default implementation switch, using `client-go` for ConfigMap reading and watch; simultaneously retains fake client injection, close-refuse-watch, native As/Underlying and behavior tests.

## Current Production Prerequisites

- `nacos` default implementation no longer uses self-built HTTP client, but reuses official `nacos-sdk-go/v2`; public contract remains unchanged; users needing vendor capabilities can down-dive to native client through As/Underlying.
- `apollo` default implementation no longer uses self-built HTTP pull, but reuses official `agollo/v4`; public contract remains unchanged; users needing Apollo native capabilities can down-dive to native client through As/Underlying.
- `apollo` current Watch has clearly distinguished "retryable errors" and "must-stop errors": SourceUnavailable will backoff-retry, AuthFailed/ConfigNotFound will no longer blindly retry.
- `consul / etcd / kubernetes` currently also have unified As/Underlying escape hatch; users can continue through gorp contract for default main path, or get underlying official Go client as needed.
- `apollo / polaris` currently both explicitly do not support Set productization write-back; if write path is needed, must independently complete full publish semantics and failure strategy.
- `apollo / polaris` currently both have fake client behavior tests, and default implementations are switched to official SDK; but still haven't completed full publish semantics, server-side long-polling detail exposure, write-back capabilities and deeper governance semantics.
- `apollo / polaris` current Watch has already reached "when not explicitly Load'ed, still pulls initial snapshot first", and when source is unreachable, retries by backoff interval, and does not re-distribute on same revision.
- `polaris` additionally has based on PollInterval fallback refresh: even if SDK watch does not actively push, can still periodically pull config and incrementally distribute based on revision.

## P0 Constraints

- Do not claim "config center capability completed" just because directory exists.
- Do not confuse `configsource/nacos` and `registry/nacos` as "full Nacos capability completed".
- Do not mistakenly write Kubernetes deployment templates as `configsource/kubernetes` completed.
