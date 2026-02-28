import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import React from "react";
import { createRoot } from "react-dom/client";
import { act } from "react";

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

vi.mock("next-export-i18n", () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-bootstrap", () => {
  const Button = ({ children, onClick, disabled, variant, size }: any) =>
    React.createElement(
      "button",
      { "data-variant": variant, "data-size": size, onClick, disabled },
      children,
    );
  const ListGroup = ({ children, className }: any) =>
    React.createElement("ul", { className }, children);
  ListGroup.Item = ({ children, className }: any) =>
    React.createElement("li", { className }, children);
  const InputGroup = ({ children, className }: any) =>
    React.createElement(
      "div",
      { "data-testid": "input-group", className },
      children,
    );
  const Form = {
    Control: ({ placeholder, value, onChange, disabled, maxLength }: any) =>
      React.createElement("input", {
        placeholder,
        value,
        onChange,
        disabled,
        maxLength,
      }),
  };
  return { Button, ListGroup, InputGroup, Form };
});

vi.mock("@/util/Formatting", () => ({
  default: {
    decodeHtmlEntities: (s: string) => s,
    getFormatterShort: () => ({ format: (d: Date) => d.toISOString() }),
  },
}));

// ---------------------------------------------------------------------------
// Passkey API mock (list returns empty by default; overridden per test)
// ---------------------------------------------------------------------------
const mockList = vi.fn().mockResolvedValue([]);
const mockBeginRegistration = vi.fn();
const mockFinishRegistration = vi.fn();
const mockIsSupported = vi.fn().mockReturnValue(true);
const mockIsPlatformAuthAvailable = vi.fn().mockResolvedValue(true);

vi.mock("@/types/Passkey", () => ({
  default: {
    list: (...args: any[]) => mockList(...args),
    beginRegistration: (...args: any[]) => mockBeginRegistration(...args),
    finishRegistration: (...args: any[]) => mockFinishRegistration(...args),
    deletePasskey: vi.fn().mockResolvedValue(undefined),
    isSupported: (...args: any[]) => mockIsSupported(...args),
    isPlatformAuthAvailable: (...args: any[]) =>
      mockIsPlatformAuthAvailable(...args),
  },
  prepareCreationOptions: (o: any) => o,
  serializeAttestationResponse: (c: any) => c,
}));

// ---------------------------------------------------------------------------
// RuntimeConfig mock
// ---------------------------------------------------------------------------
vi.mock("../RuntimeConfig", () => ({
  default: {
    INFOS: {
      isPrimaryDomain: true,
      orgPrimaryDomain: "primary.example.com",
    },
  },
}));

// Import component AFTER all mocks are set up
import PasskeySettings from "../PasskeySettings";

import RuntimeConfig from "../RuntimeConfig";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function renderComponent(props: Record<string, any> = {}): HTMLDivElement {
  const div = document.createElement("div");
  document.body.appendChild(div);
  act(() => {
    createRoot(div).render(
      React.createElement(PasskeySettings as any, {
        t: (k: string) => k,
        ...props,
      }),
    );
  });
  return div;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("PasskeySettings – primary domain gating", () => {
  let container: HTMLDivElement;

  beforeEach(() => {
    mockList.mockResolvedValue([]);
    mockIsSupported.mockReturnValue(true);
    mockIsPlatformAuthAvailable.mockResolvedValue(true);
    // Default: on primary domain
    (RuntimeConfig.INFOS as any).isPrimaryDomain = true;
    (RuntimeConfig.INFOS as any).orgPrimaryDomain = "primary.example.com";
  });

  afterEach(() => {
    if (container && container.parentNode) {
      document.body.removeChild(container);
    }
  });

  it("shows Add Passkey input and button when on primary domain", async () => {
    container = renderComponent();
    // Allow async loadPasskeys to resolve
    await act(async () => {});

    const inputGroup = container.querySelector("[data-testid='input-group']");
    expect(inputGroup).not.toBeNull();
  });

  it("shows non-primary-domain message when not on primary domain", async () => {
    (RuntimeConfig.INFOS as any).isPrimaryDomain = false;
    container = renderComponent();
    await act(async () => {});

    // The registration input group must not be rendered
    const inputGroup = container.querySelector("[data-testid='input-group']");
    expect(inputGroup).toBeNull();

    // The informational message key must be rendered
    expect(container.textContent).toContain("passkeyRegOnlyOnPrimaryDomain");
  });

  it("includes a link to the primary domain's preferences page when not on primary domain", async () => {
    (RuntimeConfig.INFOS as any).isPrimaryDomain = false;
    (RuntimeConfig.INFOS as any).orgPrimaryDomain = "primary.example.com";
    container = renderComponent();
    await act(async () => {});

    const link = container.querySelector("a[href*='primary.example.com']");
    expect(link).not.toBeNull();
    expect((link as HTMLAnchorElement).href).toContain("/ui/preferences");
  });

  it("shows the primary domain name in the link text", async () => {
    (RuntimeConfig.INFOS as any).isPrimaryDomain = false;
    (RuntimeConfig.INFOS as any).orgPrimaryDomain = "primary.example.com";
    container = renderComponent();
    await act(async () => {});

    const link = container.querySelector("a");
    expect(link?.textContent).toBe("primary.example.com");
  });

  it("still renders the passkey list (read-only) when not on primary domain", async () => {
    (RuntimeConfig.INFOS as any).isPrimaryDomain = false;
    mockList.mockResolvedValue([
      {
        id: "pk-1",
        name: "My Key",
        createdAt: new Date().toISOString(),
        lastUsedAt: null,
      },
    ]);
    container = renderComponent();
    await act(async () => {});

    // Passkey name must be visible
    expect(container.textContent).toContain("My Key");
    // Delete button must still be present
    const buttons = container.querySelectorAll("button");
    const deleteBtn = Array.from(buttons).find(
      (b) => b.textContent === "delete",
    );
    expect(deleteBtn).not.toBeUndefined();
  });

  it("does not render when platform authenticator is not available", async () => {
    mockIsSupported.mockReturnValue(false);
    mockIsPlatformAuthAvailable.mockResolvedValue(false);
    container = renderComponent();
    await act(async () => {});

    expect(container.textContent).toBe("");
  });
});
