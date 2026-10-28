import { describe, it, expect, afterEach } from "vitest";
import React, { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import OrgAnnouncement from "../OrgAnnouncement";

let div: HTMLDivElement;
let root: Root;

function render(text: string) {
  div = document.createElement("div");
  document.body.appendChild(div);
  root = createRoot(div);
  act(() => {
    root.render(<OrgAnnouncement text={text} />);
  });
}

describe("OrgAnnouncement", () => {
  afterEach(() => {
    act(() => root.unmount());
    div.remove();
  });

  it("renders nothing for empty text", () => {
    render("  \n ");
    expect(div.innerHTML).toBe("");
  });

  it("renders markdown", () => {
    render("# Title\n\nSome **bold** text and a [link](https://example.com)");
    expect(div.querySelector("h1")?.textContent).toBe("Title");
    expect(div.querySelector("strong")?.textContent).toBe("bold");
    const a = div.querySelector("a");
    expect(a?.getAttribute("href")).toBe("https://example.com");
    expect(a?.getAttribute("target")).toBe("_blank");
  });

  it("opens relative links in the same tab", () => {
    render("[Settings](/ui/admin/settings/) and [mail](mailto:a@b.c)");
    const links = div.querySelectorAll("a");
    expect(links[0].getAttribute("href")).toBe("/ui/admin/settings/");
    expect(links[0].hasAttribute("target")).toBe(false);
    expect(links[1].getAttribute("target")).toBe("_blank");
  });

  it("does not render raw HTML", () => {
    render('<script>alert(1)</script><b onclick="x">hi</b>');
    expect(div.querySelector("script")).toBeNull();
    expect(div.querySelector("b")).toBeNull();
  });
});
