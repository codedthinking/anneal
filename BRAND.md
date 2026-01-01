# anneal — Brand Guide

**Version 1.0** | January 2026
Part of **the9x.ac**
Tagline: *Cool the inbox*

---

## Brand Essence

**anneal** is a personal, terminal-first email client designed to reduce inbox load without demanding completion. It reframes email as a noisy optimization problem and rewards local, irreversible progress rather than inbox zero.

**Mental model:** simulated annealing for email.
**Promise:** steady reduction under uncertainty, no guilt.

---

## Positioning

- **Audience:** Economists, researchers, technical users
- **Use case:** Short, focused sessions to reduce backlog
- **Philosophy:** Direction over destination
- **Scope:** Personal tool first; public release later
- **Email backend:** Fastmail via JMAP (opinionated by design)

---

## Voice & Tone

- **Precise** — no motivational fluff
- **Calm** — progress without pressure
- **Insider** — optimization-native language
- **Minimal** — say less, mean more

**Do**
- "Local progress is sufficient."
- "Cooling reduces volatility."

**Avoid**
- "Inbox zero"
- "Finish everything"
- "Productivity hacks"

---

## Naming & Usage

- **Product name:** anneal (always lowercase)
- **Binary:** `anneal`
- **Spoken:** "anneal" (rhymes with email)
- **Attribution:** "by the9x.ac" on public surfaces

**Correct**
- `anneal`
- `anneal now`
- `anneal 15`

**Incorrect**
- Anneal
- ANNEAL
- anneal-mail

---

## Core Metaphors

- **Temperature:** inbox volatility
- **Cooling:** repeated short sessions
- **Acceptance:** decisive handling of messages
- **Stability:** reduced backlog variance, not zero count

No metaphor should be explained in the UI. The system implies it.

---

## UX Principles (TUI)

1. **Session-scoped success**
   - Success is defined per session, never globally.
2. **Directional feedback**
   - Show deltas, not totals.
3. **No red states**
   - No failure indicators, ever.
4. **Irreversibility**
   - Actions feel final and satisfying.
5. **Silence by default**
   - Minimal chrome, minimal copy.

---

## Metrics (Allowed)

- Net change during session (e.g. `−11`)
- Messages handled
- Time spent

**Forbidden**
- Inbox size
- "Remaining"
- Daily targets
- Streaks

---

## Visual Identity

### Color Context

anneal inherits **the9x.ac** colors.

- **Background:** `#1d1d40`
- **Primary text:** `#d4d2e3`
- **Secondary text:** `#9795b5`
- **Accent:** `#e61e25` (used sparingly)

No gradients in text. No bright success colors.

### Typography

- **Primary:** Inter
- **Weights:** 400, 600
- **Style:** compact, terminal-friendly

---

## UI Copy Guidelines

- Use verbs, not nouns.
- Prefer lowercase.
- Avoid exclamation marks.
- One idea per line.

**Examples**
- "cooling"
- "stable"
- "session complete"
- "done for now"

---

## Screens & Layout

- Single primary pane
- Fixed header with tool name
- Optional status line (right-aligned)
- No sidebars
- No badges

---

## Do's and Don'ts

### Do
- Encourage stopping early
- Make partial progress feel complete
- Default to calm states
- Keep commands short

### Don't
- Mention inbox zero
- Surface global counts
- Use progress bars to completion
- Gamify with points or streaks

---

## Legal & Attribution

- Copyright © the9x.ac
- Opinionated support: Fastmail JMAP only
- No implied endorsement by Fastmail

---

## One-line Description

**anneal**
email triage under noise
by the9x.ac

---

*This document is intentionally minimal. Changes should bias toward less UI, less copy, and fewer metrics.*
