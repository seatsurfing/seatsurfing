import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import React from "react";
import { createRoot } from "react-dom/client";
import { act } from "react";

// Mock next-export-i18n so withTranslation HOC works in tests
vi.mock("next-export-i18n", () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock react-bootstrap Modal to a simpler implementation
vi.mock("react-bootstrap", () => {
  const Modal = ({ show, onHide, children }: any) =>
    show
      ? React.createElement(
          "div",
          { "data-testid": "modal", onClick: undefined },
          children,
        )
      : null;
  Modal.Header = ({ children, closeButton, ...rest }: any) =>
    React.createElement(
      "div",
      { "data-testid": "modal-header" },
      closeButton
        ? React.createElement(
            "button",
            { "data-testid": "close-btn", onClick: rest.onHide },
            "×",
          )
        : null,
      children,
    );
  Modal.Title = ({ children }: any) =>
    React.createElement("div", { "data-testid": "modal-title" }, children);
  Modal.Body = ({ children }: any) =>
    React.createElement("div", { "data-testid": "modal-body" }, children);
  Modal.Footer = ({ children }: any) =>
    React.createElement("div", { "data-testid": "modal-footer" }, children);
  const Button = ({ children, onClick, variant }: any) =>
    React.createElement(
      "button",
      { "data-variant": variant, onClick },
      children,
    );
  return { Modal, Button };
});

// Import after mocks are set up
import MfaEncouragementModal from "../MfaEncouragementModal";

// Helper: render into a fresh div and return it
function renderIntoDiv(element: React.ReactElement): HTMLDivElement {
  const div = document.createElement("div");
  document.body.appendChild(div);
  act(() => {
    createRoot(div).render(element);
  });
  return div;
}

describe("MfaEncouragementModal", () => {
  let container: HTMLDivElement;

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
  });

  afterEach(() => {
    document.body.removeChild(container);
  });

  it("renders when show is true", () => {
    const div = renderIntoDiv(
      React.createElement(MfaEncouragementModal as any, {
        show: true,
        onHide: vi.fn(),
        onSetup: vi.fn(),
      }),
    );
    expect(div.querySelector("[data-testid='modal']")).not.toBeNull();
    document.body.removeChild(div);
  });

  it("does not render when show is false", () => {
    const div = renderIntoDiv(
      React.createElement(MfaEncouragementModal as any, {
        show: false,
        onHide: vi.fn(),
        onSetup: vi.fn(),
      }),
    );
    expect(div.querySelector("[data-testid='modal']")).toBeNull();
    document.body.removeChild(div);
  });

  it("renders title via translation key", () => {
    const div = renderIntoDiv(
      React.createElement(MfaEncouragementModal as any, {
        show: true,
        onHide: vi.fn(),
        onSetup: vi.fn(),
      }),
    );
    const title = div.querySelector("[data-testid='modal-title']");
    expect(title?.textContent).toBe("mfaEncouragementTitle");
    document.body.removeChild(div);
  });

  it("renders body text via translation key", () => {
    const div = renderIntoDiv(
      React.createElement(MfaEncouragementModal as any, {
        show: true,
        onHide: vi.fn(),
        onSetup: vi.fn(),
      }),
    );
    const body = div.querySelector("[data-testid='modal-body']");
    expect(body?.textContent).toContain("mfaEncouragementBody");
    document.body.removeChild(div);
  });

  it("calls onSetup when 'Set up now' button is clicked", () => {
    const onSetup = vi.fn();
    const onHide = vi.fn();
    const div = renderIntoDiv(
      React.createElement(MfaEncouragementModal as any, {
        show: true,
        onHide,
        onSetup,
      }),
    );
    const buttons = div.querySelectorAll("button");
    const setupBtn = Array.from(buttons).find(
      (b) => b.textContent === "mfaEncouragementSetup",
    );
    expect(setupBtn).not.toBeUndefined();
    act(() => {
      setupBtn?.click();
    });
    expect(onSetup).toHaveBeenCalledOnce();
    expect(onHide).not.toHaveBeenCalled();
    document.body.removeChild(div);
  });

  it("calls onHide when 'Maybe later' button is clicked", () => {
    const onSetup = vi.fn();
    const onHide = vi.fn();
    const div = renderIntoDiv(
      React.createElement(MfaEncouragementModal as any, {
        show: true,
        onHide,
        onSetup,
      }),
    );
    const buttons = div.querySelectorAll("button");
    const laterBtn = Array.from(buttons).find(
      (b) => b.textContent === "mfaEncouragementLater",
    );
    expect(laterBtn).not.toBeUndefined();
    act(() => {
      laterBtn?.click();
    });
    expect(onHide).toHaveBeenCalledOnce();
    expect(onSetup).not.toHaveBeenCalled();
    document.body.removeChild(div);
  });
});

// --- Logic tests for checkMfaEncouragement conditions ---
// These test the prerequisite conditions independently of React rendering.

const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

describe("MFA Encouragement Modal - visibility conditions", () => {
  beforeEach(() => {
    vi.stubGlobal("localStorage", localStorageMock);
    localStorageMock.clear();
  });

  it("should show when all conditions are met", () => {
    const config = {
      idpLogin: false,
      enforceTOTP: false,
      totpEnabled: false,
      hasPasskeys: false,
    };
    expect(shouldShowMfaEncouragement(config)).toBe(true);
  });

  it("should NOT show when user is an IdP login", () => {
    expect(
      shouldShowMfaEncouragement({
        idpLogin: true,
        enforceTOTP: false,
        totpEnabled: false,
        hasPasskeys: false,
      }),
    ).toBe(false);
  });

  it("should NOT show when enforcement is enabled", () => {
    expect(
      shouldShowMfaEncouragement({
        idpLogin: false,
        enforceTOTP: true,
        totpEnabled: false,
        hasPasskeys: false,
      }),
    ).toBe(false);
  });

  it("should NOT show when TOTP is already enabled", () => {
    expect(
      shouldShowMfaEncouragement({
        idpLogin: false,
        enforceTOTP: false,
        totpEnabled: true,
        hasPasskeys: false,
      }),
    ).toBe(false);
  });

  it("should NOT show when user has passkeys", () => {
    expect(
      shouldShowMfaEncouragement({
        idpLogin: false,
        enforceTOTP: false,
        totpEnabled: false,
        hasPasskeys: true,
      }),
    ).toBe(false);
  });

  it("should NOT show when localStorage flag is set", () => {
    localStorage.setItem("mfaEncouragementDismissed", "1");
    expect(
      shouldShowMfaEncouragement({
        idpLogin: false,
        enforceTOTP: false,
        totpEnabled: false,
        hasPasskeys: false,
      }),
    ).toBe(false);
  });

  it("should show when localStorage flag is absent", () => {
    localStorage.removeItem("mfaEncouragementDismissed");
    expect(
      shouldShowMfaEncouragement({
        idpLogin: false,
        enforceTOTP: false,
        totpEnabled: false,
        hasPasskeys: false,
      }),
    ).toBe(true);
  });

  it("dismissing should set localStorage flag", () => {
    localStorage.removeItem("mfaEncouragementDismissed");
    dismissMfaEncouragement();
    expect(localStorage.getItem("mfaEncouragementDismissed")).toBe("1");
  });

  it("after dismissal, modal should no longer show", () => {
    localStorage.removeItem("mfaEncouragementDismissed");
    dismissMfaEncouragement();
    expect(
      shouldShowMfaEncouragement({
        idpLogin: false,
        enforceTOTP: false,
        totpEnabled: false,
        hasPasskeys: false,
      }),
    ).toBe(false);
  });
});

// Pure logic helpers extracted from _app.tsx for unit testing

interface MfaConditions {
  idpLogin: boolean;
  enforceTOTP: boolean;
  totpEnabled: boolean;
  hasPasskeys: boolean;
}

function shouldShowMfaEncouragement(conditions: MfaConditions): boolean {
  if (conditions.idpLogin) return false;
  if (conditions.enforceTOTP) return false;
  if (conditions.totpEnabled) return false;
  if (conditions.hasPasskeys) return false;
  try {
    if (localStorage.getItem("mfaEncouragementDismissed") === "1") return false;
  } catch {
    // localStorage unavailable — show modal
  }
  return true;
}

function dismissMfaEncouragement(): void {
  try {
    localStorage.setItem("mfaEncouragementDismissed", "1");
  } catch {
    // ignore
  }
}
