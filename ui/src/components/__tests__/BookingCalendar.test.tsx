import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import React from "react";
import { createRoot } from "react-dom/client";
import { act } from "react";
import { getCalendarRange } from "../calendar/BookingCalendar";
import CustomToolbar from "../calendar/CustomToolbar";
import RuntimeConfig from "../RuntimeConfig";

describe("getCalendarRange", () => {
  beforeEach(() => {
    RuntimeConfig.INFOS.weekStartDay = 1;
  });

  it("returns the full week for the week view", () => {
    // Wednesday
    const date = new Date(Date.UTC(2030, 8, 4, 12, 0, 0));
    const range = getCalendarRange(date, "week");

    expect(range.start.toISOString()).toBe("2030-09-02T00:00:00.000Z");
    expect(range.end.toISOString()).toBe("2030-09-08T23:59:59.999Z");
  });

  it("includes adjacent days shown in the month view", () => {
    const date = new Date(Date.UTC(2030, 8, 15, 12, 0, 0));
    const range = getCalendarRange(date, "month");

    expect(range.start.toISOString()).toBe("2030-08-26T00:00:00.000Z");
    expect(range.end.toISOString()).toBe("2030-10-06T23:59:59.999Z");
  });
});

describe("CustomToolbar", () => {
  let container: HTMLDivElement;
  const t = (key: string) => key;

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
  });

  afterEach(() => {
    container.remove();
  });

  const render = (props: object) => {
    const root = createRoot(container);
    act(() => {
      root.render(React.createElement(CustomToolbar, props as any));
    });
    return root;
  };

  it("renders a view switch and changes the view", () => {
    const onView = vi.fn();
    render({
      t,
      views: ["week", "month"],
      toolbar: {
        date: new Date(Date.UTC(2030, 8, 15)),
        view: "week",
        onNavigate: vi.fn(),
        onView,
      },
    });

    const monthButton = Array.from(container.querySelectorAll("button")).find(
      (b) => b.textContent === "month",
    );
    expect(monthButton).toBeDefined();
    act(() => {
      monthButton!.click();
    });
    expect(onView).toHaveBeenCalledWith("month");
  });

  it("renders no view switch for a single view", () => {
    render({
      t,
      views: ["week"],
      toolbar: {
        date: new Date(Date.UTC(2030, 8, 15)),
        view: "week",
        onNavigate: vi.fn(),
        onView: vi.fn(),
      },
    });

    const labels = Array.from(container.querySelectorAll("button")).map(
      (b) => b.textContent,
    );
    expect(labels).not.toContain("month");
  });

  it("hides booking navigation without events", () => {
    render({
      t,
      toolbar: {
        date: new Date(Date.UTC(2030, 8, 15)),
        view: "week",
        onNavigate: vi.fn(),
        onView: vi.fn(),
      },
    });

    expect(container.textContent).not.toContain("nextBooking");
  });
});
