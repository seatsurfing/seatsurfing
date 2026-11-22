import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import Ajax from "./Ajax";
import AjaxError from "./AjaxError";
import User from "@/types/User";
import Role from "@/types/Role";
import Group from "@/types/Group";
import Booking from "@/types/Booking";
import Location from "@/types/Location";
import Buddy from "@/types/Buddy";
import AuthProvider from "@/types/AuthProvider";
import SpaceAttribute from "@/types/SpaceAttribute";
import RecurringBooking from "@/types/RecurringBooking";

// A URL id like "..\..\..\user\totp\disable" must never be able to escape the
// intended request path (browsers treat "\" like "/" and collapse "..").
const EVIL = "..\\..\\..\\user\\totp\\disable";
const ENCODED = encodeURIComponent(EVIL);

describe("entity helpers encode ids placed into request paths", () => {
  let calls: string[];

  beforeEach(() => {
    calls = [];
    const record =
      (result: any) =>
      async (url: string): Promise<any> => {
        calls.push(url);
        return result;
      };
    const ok = { json: {}, status: 200, objectId: "" };
    vi.spyOn(Ajax, "get").mockImplementation(record(ok));
    vi.spyOn(Ajax, "delete").mockImplementation(record(ok));
    vi.spyOn(Ajax, "putData").mockImplementation(record(ok));
    vi.spyOn(Ajax, "postData").mockImplementation(record(ok));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("does not leak raw separators for any covered call", async () => {
    await User.adminResetTotp(EVIL).catch(() => {});
    await Role.getUserIds(EVIL).catch(() => {});

    const location = new Location();
    location.id = EVIL;
    await location.getAttributes().catch(() => {});
    const buddy = new Buddy();
    buddy.id = EVIL;
    await buddy.delete().catch(() => {});

    const user = new User();
    user.id = EVIL;
    await user.delete().catch(() => {});
    await user.setPassword("x").catch(() => {});

    expect(calls.length).toBeGreaterThanOrEqual(6);
    for (const url of calls) {
      expect(url).not.toContain("\\");
      expect(url).toContain(ENCODED);
    }
  });

  it("encodes the id on saveEntity updates", async () => {
    const e: any = { id: EVIL, serialize: () => ({}) };
    await Ajax.saveEntity(e, "/user/");
    expect(calls).toEqual(["/user/" + ENCODED]);
  });
});

describe("Ajax.path", () => {
  it("encodes interpolated values", () => {
    expect(Ajax.path`/user/${"a b+c"}/roles`).toBe("/user/a%20b%2Bc/roles");
  });

  it("returns static paths unchanged", () => {
    expect(Ajax.path`/user/me`).toBe("/user/me");
  });

  it.each([[""], ["."], [".."], ["../auth/logout/all"], ["a/b"], [EVIL]])(
    "rejects %j with a 404",
    (value) => {
      expect(() => Ajax.path`/user/${value}`).toThrow(AjaxError);
    },
  );
});

describe("getters reject traversal ids without sending a request", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it.each<[string, (id: string) => Promise<unknown>]>([
    ["User", (id) => User.get(id)],
    ["Role", (id) => Role.get(id)],
    ["Group", (id) => Group.get(id)],
    ["Booking", (id) => Booking.get(id)],
    ["Location", (id) => Location.get(id)],
    ["AuthProvider", (id) => AuthProvider.get(id)],
    ["SpaceAttribute", (id) => SpaceAttribute.get(id)],
    ["RecurringBooking", (id) => RecurringBooking.get(id)],
  ])("%s.get", async (_name, call) => {
    const get = vi.spyOn(Ajax, "get");
    for (const id of ["", ".", "..", "../auth/logout/all", EVIL]) {
      await expect(call(id)).rejects.toBeInstanceOf(AjaxError);
    }
    expect(get).not.toHaveBeenCalled();
  });
});
