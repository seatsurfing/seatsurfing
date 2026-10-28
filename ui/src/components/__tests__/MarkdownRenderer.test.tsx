import { describe, it, expect, afterEach } from "vitest";
import React, { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import MarkdownRenderer from "../MarkdownRenderer";

let div: HTMLDivElement;
let root: Root;

function render(element: React.ReactElement) {
  div = document.createElement("div");
  document.body.appendChild(div);
  root = createRoot(div);
  act(() => {
    root.render(element);
  });
}

const md = "[rel](/ui/admin/) [ext](https://example.com) [mail](mailto:a@b.c)";

describe("MarkdownRenderer", () => {
  afterEach(() => {
    act(() => root.unmount());
    div.remove();
  });

  it("opens all links in a new tab by default", () => {
    render(<MarkdownRenderer>{md}</MarkdownRenderer>);
    div.querySelectorAll("a").forEach((a) => {
      expect(a.getAttribute("target")).toBe("_blank");
    });
  });

  it("opens relative links in the same tab if enabled", () => {
    render(<MarkdownRenderer relativeLinksSameTab>{md}</MarkdownRenderer>);
    const links = div.querySelectorAll("a");
    expect(links[0].hasAttribute("target")).toBe(false);
    expect(links[1].getAttribute("target")).toBe("_blank");
    expect(links[2].getAttribute("target")).toBe("_blank");
  });
});
