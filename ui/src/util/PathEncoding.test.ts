import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import Ajax from "./Ajax";
import User from "@/types/User";
import Role from "@/types/Role";
import Group from "@/types/Group";
import Booking from "@/types/Booking";
import Location from "@/types/Location";
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
    await User.get(EVIL).catch(() => {});
    await Role.get(EVIL).catch(() => {});
    await Group.get(EVIL).catch(() => {});
    await Booking.get(EVIL).catch(() => {});
    await Location.get(EVIL).catch(() => {});
    await AuthProvider.get(EVIL).catch(() => {});
    await SpaceAttribute.get(EVIL).catch(() => {});
    await RecurringBooking.get(EVIL).catch(() => {});
    await User.adminResetTotp(EVIL).catch(() => {});
    await Role.getUserIds(EVIL).catch(() => {});

    const user = new User();
    user.id = EVIL;
    await user.delete().catch(() => {});
    await user.setPassword("x").catch(() => {});

    expect(calls.length).toBeGreaterThanOrEqual(12);
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
