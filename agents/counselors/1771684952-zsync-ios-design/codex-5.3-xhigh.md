**Overall Verdict**
Pure-native Swift is the right call. The design is strong, but there is one major blocker: `mosh` on iOS without GPL contamination. Everything else is solvable with good scoping.

**1) Architecture**
- Yes: `SwiftUI + SwiftTerm + SwiftData/CloudKit + Keychain` is a sensible v1 stack.
- Keep `Citadel` behind your own transport protocol so you can swap to raw `swift-nio-ssh` if needed.
- Biggest stack caveat: `swift-nio-ssh` is a low-level toolkit, not a full SSH client; you must own PTY, half-close behavior, channel lifecycle, host key UX, etc.
- `SwiftTerm` is mature and active, but it still calls out selection/accessibility caveats. Plan explicit accessibility work.

**2) UX**
- Three-screen flow is correct for iPhone.
- Quick-jump recents are good, but add stale-state handling: session missing, machine offline, renamed session.
- Composition + key bar is directionally right, but the Passthrough/Compose mode toggle can cause mode errors. Prefer a mode-light design:
  - Default passthrough.
  - “Compose once” action instead of global mode switch.
- Swipe-to-kill is risky for destructive remote actions. Require confirm or undo window.

**3) Biggest Technical Risks (ranked)**
1. iOS background behavior. You cannot rely on persistent UDP/TCP when suspended; treat resume as reconnect-first, not continuous.
2. Host key trust + SSH security UX. A weak TOFU flow will burn trust fast.
3. Terminal fidelity under mobile input. External keyboard, modifier-state, IME/dictation, and paste behavior are where bugs surface.
4. Session discovery latency. SSH-ing into multiple machines serially will feel slow; use parallel fetch + cache + timeout.
5. Cloud sync edge cases. SwiftData+CloudKit has real pitfalls around schema/default values and iCloud-toggle behavior.

**4) Missing V1 Features**
- Known-hosts management (fingerprint review, pinning, rotation handling).
- Onboarding health check per machine (`zmosh` installed, UDP reachable, auth valid).
- Session search/filter (not just list).
- Export/import backup for machine configs.
- External keyboard optimized mode (hide key bar, richer shortcuts).
- Biometric app lock and per-key usage policy.
- Failure visibility: reconnect reason, auth failures, machine health badges.

**5) Licensing (critical)**
- If you link GPL-3.0 `libmoshios`/Blink code, you likely pull your app into GPL obligations. That conflicts with your desired proprietary/commercial path.
- Recommendation: do **not** ship GPL mosh client code in v1.
- Safer options:
  - Ship SSH-first v1.
  - Add clean-room mosh-compatible implementation later.
  - Consider Eternal Terminal support (Apache-2.0) as resilient-session alternative.
- Not legal advice, but this needs real counsel before implementation.

**6) Voice Input**
- Keyboard dictation alone is enough for v1.
- Custom `SFSpeechRecognizer` language model is useful later, but only after core terminal reliability is done.
- If you add custom voice:
  - Compose-only.
  - Never auto-send to shell.
  - Mandatory preview/confirm.
- Speech APIs also have operational limits/permissions overhead that add support burden.

**7) Sync**
- SwiftData + CloudKit is fine for machine metadata.
- Don’t assume key material sync “just works.” Decide explicitly what syncs.
- Keep private keys local by default; only sync if intentionally marked synchronizable and policy-approved.
- Strong recommendation: separate local source-of-truth from cloud mirror semantics to avoid destructive surprises when iCloud state changes.

**8) Monetization**
As of **February 21, 2026**, market anchor points are high (`Termius` Pro ~$15/mo or $119/yr; `Prompt 3` ~$19.99/yr or $99.99 one-time).  
Best path for your niche:
- Free tier: 1 machine, SSH only.
- One-time Pro unlock: unlimited machines, advanced keyboard/theming, recents/power UX.
- Optional annual add-on: premium sync/collab/AI voice extras.
- Avoid subscription-only at launch; terminal users strongly value ownership.

**Recommended scope adjustment**
- V1: SSH + zmosh picker workflow + polished keyboard + secure key/host trust + fast discovery.
- V1.5: Cloud polish + external keyboard depth.
- V2: mosh-compatible transport only after licensing-safe path is settled.

**Sources**
- Local project context: `README.md`, `docs/plans/2026-02-20-zmosh-picker-design.md`, `zmosh-picker`, `prototype/zsync.html`
- SwiftTerm: https://github.com/migueldeicaza/SwiftTerm
- Citadel: https://github.com/orlandos-nl/Citadel
- SwiftNIO SSH: https://github.com/apple/swift-nio-ssh
- Blink repo/license context: https://github.com/blinksh/blink
- Mosh license: https://github.com/mobile-shell/mosh
- GNU GPL FAQ: https://www.gnu.org/licenses/gpl-faq.html
- Apple background behavior docs: https://developer.apple.com/library/archive/documentation/Performance/Conceptual/EnergyGuide-iOS/WorkLessInTheBackground.html
- Speech framework docs/sample: https://developer.apple.com/documentation/speech/recognizing-speech-in-live-audio
- Speech API limits QA1951: https://developer.apple.com/library/archive/qa/qa1951/_index.html
- CloudKit limits/errors: https://developer.apple.com/library/archive/documentation/DataManagement/Conceptual/CloudKitWebServicesReference/PropertyMetrics.html
- SwiftData/CloudKit DTS forum notes: https://developer.apple.com/forums/thread/779302 and https://developer.apple.com/forums/thread/805940
- App Store export compliance: https://developer.apple.com/help/app-store-connect/manage-app-information/overview-of-export-compliance
- Competitor pricing/features: https://apps.apple.com/us/app/termius/id549039908 and https://apps.apple.com/us/app/prompt-3/id1594420480
