# Design Review: Zsync for iOS

## Executive Summary
**Verdict:** **High Potential / High Technical Risk**

The concept of "cmux for iOS" is excellent. The mobile SSH experience is generally poor because of connection fragility and keyboard limitations. Zsync solves both by offloading persistence to the server (`zmosh`) and focusing the client on *reconnection* rather than *maintenance*.

However, the reliance on `libmoshios` (GPL-3.0) forces a licensing decision immediately, and the "pure Swift" goal hits a wall with Mosh's complex C++ protocol.

## 1. Architecture & Tech Stack

### Pure Native Swift (SwiftUI + SwiftTerm)
*   **Verdict:** **Correct Choice.**
*   **Analysis:** Avoid React Native or Flutter for terminal apps. Text rendering performance and keyboard handling are critical; wrappers introduce lag and key-binding headaches.
*   **SwiftTerm Risk:** While Miguel de Icaza’s `SwiftTerm` is the gold standard for Swift emulation, integration is not trivial. The `SwiftTermApp` reference implementation has known issues with high-throughput SSH (data loss/rendering glitches). You will likely need to patch `SwiftTerm` or write a robust custom glue layer between `Citadel` and `SwiftTerm`.

### SSH & Mosh
*   **Citadel / swift-nio-ssh:** **Solid.** These are performant and well-maintained.
*   **libmoshios (The Blocker):**
    *   **Licensing:** `libmoshios` is GPL-3.0 (derived from Mosh). Linking it **requires your entire app to be GPL-3.0**. You can still sell it on the App Store (like Blink Shell), but you must make the source code available to customers. You cannot close-source this app if you use that library.
    *   **Implementation:** Mosh is C++. Bridging C++ Mosh to Swift requires a substantial Objective-C++ wrapper. This breaks your "Pure Swift" architectural purity but is unavoidable unless you rewrite Mosh in Swift (a massive undertaking).

## 2. UX & Workflow

### Navigation
*   **Three-Screen Flow:** The **Machines -> Sessions -> Terminal** hierarchy is correct. It mirrors the mental model of "Where am I going?" -> "What am I doing?".
*   **Quick Jump:** This is the killer feature. Most users rotate between 2-3 active contexts. This should be the default view on launch.

### Input Handling
*   **Passthrough vs. Compose Mode:** **Brilliant.** This solves the single biggest frustration of mobile SSH: latency and typos.
    *   *Suggestion:* In "Compose" mode, allow local history. If I type a long command and send it, I might want to edit it slightly and send it again. The terminal's history won't help me here because the composition input is local.
*   **Key Bar:** The specific keys chosen (`esc`, `ctrl`, `tab`, `|`, `/`, `-`, `_`) are spot on.
    *   *Missing:* Arrow keys are crucial. Your prototype has them, but ensure they work with modifiers (e.g., Ctrl+Left to jump words).

## 3. Technical Risks & Blind Spots

### The "Discovery" Fragility
*   **Risk:** The app parses the text output of `zmosh list`. If you change `zmosh`'s output format in a future update, **every installed version of the iOS app breaks**.
*   **Fix:** Update `zmosh` CLI to support a structured output flag (e.g., `zmosh list --json`) and have the app use that. Do not rely on regex parsing of user-facing text.

### Backgrounding & "Zombie" States
*   **The Problem:** iOS allows sockets to stay open for a short window, but eventually kills them.
    *   *SSH (Discovery):* When the user backgrounds the app on the "Session Picker" screen, the SSH connection used to list sessions will die. When they return, the list is stale. If they tap a session, it might fail.
    *   *Mosh (Terminal):* Mosh handles this natively via UDP roaming. This is safe.
*   **Mitigation:** The "Session Picker" screen needs aggressive "refresh on wake" logic. It cannot assume the SSH connection from 3 minutes ago is still alive.

### SSH Key Management
*   **Risk:** Users hate copying private keys to phones.
*   **Blind Spot:** How do they get the key there? AirDrop? Pasteboard? iTunes File Sharing?
*   **Recommendation:** Support **Secure Enclave** key generation. Let the phone generate a key and display the public key for the user to add to `~/.ssh/authorized_keys` on their Mac. This is more secure and easier than transferring private keys.

## 4. Missing Features for v1

1.  **Structured Output Protocol:** As mentioned, `zmosh list --json`.
2.  **Snippet Library:** Mobile typing is slow. Allow users to save "Restart Server", "Git Status", or "Deploy" as clickable buttons in the accessory bar.
3.  **External Keyboard Support:** Even though it's "iPhone First", many users will pair a bluetooth keyboard. SwiftTerm handles this well, but ensure your "Compose" bar doesn't capture input when a hardware keyboard is attached.
4.  **Agent Forwarding:** If I SSH from Phone -> Mac, and then `git push` from the Mac, I need my SSH keys. `zmosh` on the Mac handles the agent, but the initial connection might need to forward an agent if the user uses keys stored on the phone for GitHub auth.

## 5. Voice Input Review
*   **Verdict:** **High Value Differentiator.**
*   **SFSpeechRecognizer:** Essential. Standard iOS dictation is trained on conversational English. It will autocorrect `ls -la` to `let's la` or `is la`.
*   **Custom Vocabulary:** You must train the model with:
    *   Shell builtins (`ls`, `cd`, `grep`, `awk`, `sed`)
    *   Common flags (`-l`, `-a`, `-h`, `--help`)
    *   Symbols ("pipe" -> `|`, "dash" -> `-`, "slash" -> `/`)
*   **Feasibility:** Totally feasible and a great demo feature, but heavy on battery/processing. Keep it behind a dedicated button (as designed).

## 6. Sync (SwiftData + CloudKit)
*   **Verdict:** **Approved.**
*   For simple configuration data (Machine IP, Port, Username, Theme settings), this is perfect. It requires zero backend maintenance and "just works" for users in the Apple ecosystem.

## 7. Monetization Strategy
*   **The "Blink" Model:** Since you are likely forced into GPL-3.0 by `libmoshios`:
    *   **Price:** Paid upfront ($19.99) or Subscription ($19/year).
    *   **Justification:** You are selling the *compiled, signed, convenient binary* and the *CloudKit sync service*. Power users understand that maintaining an iOS dev environment to build it themselves is worth more than $20.
    *   **Free Tier:** One machine, no sync. Pro: Unlimited machines + Sync.

## Summary of Recommendations

1.  **Accept the GPL:** Use `libmoshios` (from Blink). Don't try to write a pure Swift Mosh client yet.
2.  **Update CLI first:** Add JSON output to `zmosh` before building the app's parser.
3.  **Refine "Compose" Mode:** Add local history and snippets.
4.  **Key Transfer:** Build a flow to generate keys on-device and export the public key.
