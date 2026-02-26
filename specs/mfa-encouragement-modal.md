# Feature Spec: MFA Encouragement Modal

## Overview

After logging in, users who have not yet set up a second factor (TOTP or Passkey) are shown a one-time modal encouraging them to do so. The modal is dismissible and does not block usage of the application. Once dismissed (either way), a localStorage flag prevents it from appearing again in the same browser.

## Prerequisites for Showing the Modal

All of the following conditions must be true:

1. **User is logged in** — `Ajax.hasAccessToken()` returns `true`
2. **Not on a login/auth page** — current path does not include `/login` or `/resetpw`
3. **Not an IdP user** — `RuntimeConfig.INFOS.idpLogin === false`
4. **Enforcement is NOT enabled** — `RuntimeConfig.INFOS.enforceTOTP === false`
5. **No second factor is set up** — `RuntimeConfig.INFOS.totpEnabled === false` AND `RuntimeConfig.INFOS.hasPasskeys === false`
6. **Modal has not been previously dismissed** — `localStorage.getItem("mfaEncouragementDismissed")` is not `"1"`

> **Note:** When enforcement is enabled, the existing `TotpSetupModal` with `canClose={false}` already handles forcing setup. This new modal only applies to the non-enforced "encouragement" scenario.

## Component: `MfaEncouragementModal`

### File Location

`ui/src/components/MfaEncouragementModal.tsx`

### Component Pattern

- Class component following the same pattern as `TotpSetupModal.tsx`
- Wrapped with `withTranslation` HOC for i18n support
- Uses React-Bootstrap `<Modal>` component

### Props

| Prop | Type | Description |
|------|------|-------------|
| `show` | `boolean` | Controls modal visibility |
| `onHide` | `() => void` | Called when the modal is dismissed (either button) |
| `onSetup` | `() => void` | Called when user chooses to set up MFA |
| `t` | `TranslationFunc` | Translation function (injected by `withTranslation`) |

### Modal Content

- **Title:** Translated string `mfaEncouragementTitle` (e.g. "Secure your account")
- **Body:** 
  - An informational paragraph explaining the benefit of second-factor authentication, using translated string `mfaEncouragementBody`
  - A brief mention that the user can set up either **TOTP (authenticator app)** or **Passkeys (biometrics/security key)**
- **Footer / Buttons:**
  - **Primary button:** "Set up now" (`mfaEncouragementSetup`) — calls `onSetup`
  - **Secondary/link button:** "Maybe later" (`mfaEncouragementLater`) — calls `onHide`
- The modal is closable: `canClose` behavior (close button in header, click-outside dismissal, Escape key) — all call `onHide`

### No Internal State Required

The modal is purely presentational. All logic (visibility, dismissal persistence, navigation) is handled by the parent (`_app.tsx`).

## Integration in `_app.tsx`

### New State Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `showMfaEncouragement` | `boolean` | `false` | Whether to show the encouragement modal |

### Logic Flow

Add a new method `checkMfaEncouragement()` that is called **after** the existing `checkTotpEnforcement()` completes (only when enforcement does NOT trigger). The check order is:

1. `checkTotpEnforcement()` runs first (existing behavior)
2. If enforcement modal is **not** shown, call `checkMfaEncouragement()`

`checkMfaEncouragement()` implementation:

```
1. Guard: return early if any prerequisite is not met (see Prerequisites above)
2. Read localStorage key "mfaEncouragementDismissed"
3. If key is "1", return early
4. Set state: showMfaEncouragement = true
```

### Handling User Actions

**"Set up now" (`onSetup`):**
1. Set `localStorage.setItem("mfaEncouragementDismissed", "1")`
2. Set state: `showMfaEncouragement = false`
3. Navigate to `/preferences?tab=security` using `Next.js router`

**"Maybe later" (`onHide`):**
1. Set `localStorage.setItem("mfaEncouragementDismissed", "1")`
2. Set state: `showMfaEncouragement = false`

### Rendering

Render `<MfaEncouragementModal>` in the same location as the existing `<TotpSetupModal>` block in the `render()` method — the two are mutually exclusive (enforcement modal takes priority).

```tsx
{this.state.showTotpEnforcement && (
  <TotpSetupModal ... />
)}
{this.state.showMfaEncouragement && (
  <MfaEncouragementModal
    show={true}
    onHide={this.onMfaEncouragementDismiss}
    onSetup={this.onMfaEncouragementSetup}
  />
)}
```

## Navigation Target

The "Set up now" action navigates to the preferences page with the Security tab pre-selected:

**URL:** `/preferences?tab=security`

The `preferences.tsx` page must support reading the `tab` query parameter to set `activeTab` on mount. Currently `activeTab` defaults to `"tab-bookings"` — add logic in `componentDidMount` (or constructor) to read `router.query.tab` and, if it equals `"security"`, set `activeTab` to `"tab-security"`.

## localStorage Key

| Key | Value | Description |
|-----|-------|-------------|
| `mfaEncouragementDismissed` | `"1"` | Set when the user dismisses the modal (either action). Persists across sessions. |

The key is accessed directly via `window.localStorage` (same pattern as `next-export-i18n-lang` in `RuntimeConfig.getLanguage()`). It is NOT managed through `AjaxConfigBrowserPersister` since it is unrelated to auth credentials.

The key is **not** cleared on logout. This is intentional: the encouragement is per-browser, and showing it again after re-login would be annoying. If a different user logs in on the same browser, they also won't see it — this is an acceptable tradeoff for simplicity.

## Internationalization

### New Translation Keys

Add the following keys to all 13 translation files in `ui/i18n/`:

| Key | en-US Value |
|-----|-------------|
| `mfaEncouragementTitle` | `"Secure your account"` |
| `mfaEncouragementBody` | `"Protect your account by setting up a second factor for authentication. You can use an authenticator app (TOTP) or a passkey (biometrics or security key) for additional security when signing in."` |
| `mfaEncouragementSetup` | `"Set up now"` |
| `mfaEncouragementLater` | `"Maybe later"` |

German (`de`) translations:

| Key | de Value |
|-----|----------|
| `mfaEncouragementTitle` | `"Konto absichern"` |
| `mfaEncouragementBody` | `"Schützen Sie Ihr Konto, indem Sie einen zweiten Faktor für die Authentifizierung einrichten. Sie können eine Authentifikator-App (TOTP) oder einen Passkey (Biometrie oder Sicherheitsschlüssel) verwenden, um die Sicherheit bei der Anmeldung zu erhöhen."` |
| `mfaEncouragementSetup` | `"Jetzt einrichten"` |
| `mfaEncouragementLater` | `"Vielleicht später"` |

For all other languages: add the keys with the en-US values as placeholders (consistent with existing practice — see e.g. `translations.es.json` where many keys use English text). The existing `add-missing-translations.sh` script can be used for this.

## Testing

### Unit Tests

Add tests in a new file `ui/src/components/__tests__/MfaEncouragementModal.test.tsx` (or alongside existing test patterns):

1. **Renders when `show={true}`** — modal is visible with correct title, body, and both buttons
2. **Does not render when `show={false}`** — modal is not in the DOM
3. **"Set up now" calls `onSetup`** — simulate click, verify callback
4. **"Maybe later" calls `onHide`** — simulate click, verify callback
5. **Modal is closable** — close button in header calls `onHide`

### Integration Tests (in `_app.tsx` logic)

6. **Modal shown when all prerequisites met** — mock `RuntimeConfig.INFOS` with `enforceTOTP=false`, `totpEnabled=false`, `hasPasskeys=false`, `idpLogin=false` and empty localStorage
7. **Modal NOT shown when user has TOTP** — `totpEnabled=true`
8. **Modal NOT shown when user has passkeys** — `hasPasskeys=true`
9. **Modal NOT shown when user is IdP login** — `idpLogin=true`
10. **Modal NOT shown when enforcement is enabled** — `enforceTOTP=true` (enforcement modal shown instead)
11. **Modal NOT shown when localStorage flag is set** — `mfaEncouragementDismissed = "1"`
12. **"Set up now" sets localStorage flag and navigates** — verify `localStorage.setItem` called and router navigated to `/preferences?tab=security`
13. **"Maybe later" sets localStorage flag** — verify `localStorage.setItem` called and modal hidden

### E2E Tests

Not required for initial implementation. The existing e2e framework (`e2e/tests/`) can be extended later.

## Files Changed

| File | Change |
|------|--------|
| `ui/src/components/MfaEncouragementModal.tsx` | **New** — modal component |
| `ui/src/pages/_app.tsx` | Add state, check logic, handlers, render `MfaEncouragementModal` |
| `ui/src/pages/preferences.tsx` | Support `?tab=security` query param to pre-select Security tab |
| `ui/i18n/translations.*.json` (13 files) | Add 4 new translation keys each |
| `ui/src/components/__tests__/MfaEncouragementModal.test.tsx` | **New** — unit tests |

## Edge Cases & Design Decisions

1. **Enforcement vs. Encouragement priority:** If `enforceTOTP` is enabled, only the enforcement modal is shown (blocking, non-dismissible). The encouragement modal never appears alongside or instead of enforcement.

2. **Race condition with enforcement toggle:** If an admin enables enforcement while a user has already dismissed the encouragement modal, the enforcement modal still appears correctly because it uses a separate code path (`checkTotpEnforcement`) that does not check localStorage.

3. **Multiple tabs:** If the user dismisses the modal in one tab, other already-open tabs may still show it until page reload. This is acceptable — the modal is non-blocking and harmless to see twice.

4. **localStorage not available:** If localStorage throws (e.g. private browsing in some browsers), the modal will show on every login. Wrap localStorage access in try/catch (consistent with `AjaxConfigBrowserPersister` pattern).

5. **No backend changes required:** All data needed for the prerequisites is already available in `RuntimeConfig.INFOS`, loaded from existing `/user/me` and `/setting/` endpoints.
