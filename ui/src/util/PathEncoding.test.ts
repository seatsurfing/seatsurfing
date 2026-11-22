import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import Ajax from "./Ajax";
import AjaxError from "./AjaxError";
import { getIcal } from "@/components/Ical";
import User from "@/types/User";
import Role from "@/types/Role";
import Group from "@/types/Group";
import Booking from "@/types/Booking";
import BookingAttribute from "@/types/BookingAttribute";
import Location from "@/types/Location";
import Space from "@/types/Space";
import Buddy from "@/types/Buddy";
import AuthProvider, { AuthProviderMapping } from "@/types/AuthProvider";
import SpaceAttribute from "@/types/SpaceAttribute";
import RecurringBooking from "@/types/RecurringBooking";
import Organization from "@/types/Organization";
import Domain from "@/types/Domain";
import Session from "@/types/Session";
import Passkey from "@/types/Passkey";
import Settings from "@/types/Settings";
import UserPreference from "@/types/UserPreference";

// A value like "..\..\..\user\totp\disable" or "../auth/logout/all" must never
// be able to escape the intended request path: browsers treat "\" like "/" and
// collapse "..", and the server router decodes "%2F" and cleans "../" before
// routing (so encoding alone is not enough).
const EVIL = "..\\..\\..\\user\\totp\\disable";
const BAD_VALUES = ["", ".", "..", "../auth/logout/all", "a/b", EVIL];

describe("Ajax.path", () => {
  it("encodes interpolated values", () => {
    expect(Ajax.path`/user/${"a b+c"}/roles`).toBe("/user/a%20b%2Bc/roles");
  });

  it("returns static paths unchanged", () => {
    expect(Ajax.path`/user/me`).toBe("/user/me");
  });

  it.each(BAD_VALUES.map((v) => [v]))("rejects %j with a 404", (value) => {
    expect(() => Ajax.path`/user/${value}`).toThrow(AjaxError);
  });
});

function withId<T extends { id: string }>(e: T, id: string): T {
  e.id = id;
  return e;
}

// Every method that puts an id (or other caller-supplied value) in a path.
// Each entry receives a hostile value and must fail without any request.
const CALLS: [string, (v: string) => unknown][] = [
  // User
  ["User.get", (v) => User.get(v)],
  ["User.getByEmail", (v) => User.getByEmail(v)],
  ["User.getTotpSecret", (v) => User.getTotpSecret(v)],
  ["User.adminResetPasskeys", (v) => User.adminResetPasskeys(v)],
  ["User.adminResetTotp", (v) => User.adminResetTotp(v)],
  ["User.getApiTokenStatus", (v) => User.getApiTokenStatus(v)],
  ["User.generateApiToken", (v) => User.generateApiToken(v)],
  ["User.revokeApiToken", (v) => User.revokeApiToken(v)],
  ["user.delete", (v) => withId(new User(), v).delete()],
  ["user.setPassword", (v) => withId(new User(), v).setPassword("x")],
  // Role
  ["Role.get", (v) => Role.get(v)],
  ["Role.getUserIds", (v) => Role.getUserIds(v)],
  ["Role.getRoleIdsForUser", (v) => Role.getRoleIdsForUser(v)],
  ["Role.setRoleIdsForUser", (v) => Role.setRoleIdsForUser(v, [])],
  ["role.delete", (v) => withId(new Role(), v).delete()],
  // Group
  ["Group.get", (v) => Group.get(v)],
  ["group.delete", (v) => withId(new Group(), v).delete()],
  ["group.getMembers", (v) => withId(new Group(), v).getMembers()],
  ["group.addMembers", (v) => withId(new Group(), v).addMembers([])],
  ["group.removeMembers", (v) => withId(new Group(), v).removeMembers([])],
  // Location
  ["Location.get", (v) => Location.get(v)],
  ["location.delete", (v) => withId(new Location(), v).delete()],
  ["location.getMap", (v) => withId(new Location(), v).getMap()],
  ["location.getMapUrl", (v) => withId(new Location(), v).getMapUrl()],
  [
    "location.setMap",
    (v) => withId(new Location(), v).setMap(new File([""], "m.png")),
  ],
  [
    "location.getFloorPlanDesign",
    (v) => withId(new Location(), v).getFloorPlanDesign(),
  ],
  [
    "location.setFloorPlanDesign",
    (v) => withId(new Location(), v).setFloorPlanDesign("{}"),
  ],
  ["location.getAttributes", (v) => withId(new Location(), v).getAttributes()],
  [
    "location.setAttribute (id)",
    (v) => withId(new Location(), v).setAttribute("a", "b"),
  ],
  [
    "location.setAttribute (attribute)",
    (v) => withId(new Location(), "abc").setAttribute(v, "b"),
  ],
  [
    "location.deleteAttribute (attribute)",
    (v) => withId(new Location(), "abc").deleteAttribute(v),
  ],
  // Space
  ["Space.get (location)", (v) => Space.get(v, "abc")],
  ["Space.get (space)", (v) => Space.get("abc", v)],
  ["Space.list", (v) => Space.list(v)],
  [
    "Space.listAvailability",
    (v) => Space.listAvailability(v, new Date(), new Date()),
  ],
  [
    "Space.listSingleAvailability",
    (v) => Space.listSingleAvailability("abc", v, new Date(), new Date()),
  ],
  ["Space.bulkUpdate", (v) => Space.bulkUpdate(v, [], [], [])],
  ["space.delete", (v) => withId(new Space(), v).delete()],
  ["space.getApprovers", (v) => withId(new Space(), v).getApprovers()],
  ["space.addApprovers", (v) => withId(new Space(), v).addApprovers([])],
  ["space.removeApprovers", (v) => withId(new Space(), v).removeApprovers([])],
  [
    "space.getAllowedBookers",
    (v) => withId(new Space(), v).getAllowedBookers(),
  ],
  [
    "space.addAllowedBookers",
    (v) => withId(new Space(), v).addAllowedBookers([]),
  ],
  [
    "space.removeAllowedBookers",
    (v) => withId(new Space(), v).removeAllowedBookers([]),
  ],
  ["space.getImageUrl", (v) => withId(new Space(), v).getImageUrl()],
  ["space.getImage", (v) => withId(new Space(), v).getImage()],
  ["space.deleteImage", (v) => withId(new Space(), v).deleteImage()],
  // Booking / recurring booking
  ["Booking.get", (v) => Booking.get(v)],
  ["booking.delete", (v) => withId(new Booking(), v).delete()],
  ["booking.approve", (v) => withId(new Booking(), v).approve(true)],
  ["RecurringBooking.get", (v) => RecurringBooking.get(v)],
  [
    "recurringBooking.delete",
    (v) => withId(new RecurringBooking(), v).delete(),
  ],
  // Attributes
  ["SpaceAttribute.get", (v) => SpaceAttribute.get(v)],
  ["spaceAttribute.delete", (v) => withId(new SpaceAttribute(), v).delete()],
  [
    "bookingAttribute.move",
    (v) => withId(new BookingAttribute(), v).move(true),
  ],
  [
    "bookingAttribute.delete",
    (v) => withId(new BookingAttribute(), v).delete(),
  ],
  // Auth providers
  ["AuthProvider.get", (v) => AuthProvider.get(v)],
  ["AuthProvider.listPublicForOrg", (v) => AuthProvider.listPublicForOrg(v)],
  ["authProvider.delete", (v) => withId(new AuthProvider(), v).delete()],
  ["AuthProviderMapping.list", (v) => AuthProviderMapping.list(v)],
  // Misc entities
  ["buddy.delete", (v) => withId(new Buddy(), v).delete()],
  ["session.delete", (v) => withId(new Session(), v).delete()],
  ["Passkey.rename", (v) => Passkey.rename(v, "n")],
  ["Passkey.deletePasskey", (v) => Passkey.deletePasskey(v)],
  ["Organization.get", (v) => Organization.get(v)],
  ["Organization.getOrgForDomain", (v) => Organization.getOrgForDomain(v)],
  ["organization.delete", (v) => withId(new Organization(), v).delete()],
  ["Domain.add (org)", (v) => Domain.add(v, "example.com")],
  ["Domain.add (domain)", (v) => Domain.add("abc", v)],
  ["Domain.list", (v) => Domain.list(v)],
  [
    "domain.delete",
    (v) =>
      Object.assign(new Domain(), {
        organizationId: "abc",
        domain: v,
      }).delete(),
  ],
  [
    "domain.setPrimary",
    (v) =>
      Object.assign(new Domain(), {
        organizationId: v,
        domain: "a.b",
      }).setPrimary(),
  ],
  [
    "domain.verify",
    (v) =>
      Object.assign(new Domain(), {
        organizationId: "abc",
        domain: v,
      }).verify(),
  ],
  ["Settings.getOne", (v) => Settings.getOne(v)],
  ["Settings.setOne", (v) => Settings.setOne(v, "x")],
  ["UserPreference.getOne", (v) => UserPreference.getOne(v)],
  ["UserPreference.setOne", (v) => UserPreference.setOne(v, "x")],
  // Shared helpers
  ["getIcal (booking)", (v) => getIcal(v)],
  ["getIcal (recurring)", (v) => getIcal(v, true)],
];

describe("request paths reject traversal values without a request", () => {
  let requests: unknown[];

  beforeEach(() => {
    requests = [];
    const record = async (...args: unknown[]) => {
      requests.push(args[0]);
      return { json: [], status: 200, objectId: "" };
    };
    vi.spyOn(Ajax, "get").mockImplementation(record as any);
    vi.spyOn(Ajax, "delete").mockImplementation(record as any);
    vi.spyOn(Ajax, "putData").mockImplementation(record as any);
    vi.spyOn(Ajax, "postData").mockImplementation(record as any);
    vi.spyOn(Ajax, "head").mockImplementation(record as any);
    vi.spyOn(global, "fetch").mockImplementation((async (
      ...args: unknown[]
    ) => {
      requests.push(args[0]);
      return new Response("{}");
    }) as any);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it.each(CALLS)("%s", async (_name, call) => {
    for (const value of BAD_VALUES) {
      let failure: unknown = undefined;
      try {
        await call(value);
      } catch (e) {
        failure = e;
      }
      expect(failure, `value ${JSON.stringify(value)}`).toBeInstanceOf(
        AjaxError,
      );
    }
    expect(requests).toEqual([]);
  });

  it("Ajax.saveEntity rejects traversal ids (an empty id means create)", async () => {
    for (const value of BAD_VALUES.filter((v) => v !== "")) {
      await expect(
        Ajax.saveEntity(withId(new Role(), value), "/role/"),
      ).rejects.toBeInstanceOf(AjaxError);
    }
    expect(requests).toEqual([]);
  });

  it("encodes ordinary ids", async () => {
    await withId(new User(), "a b").delete();
    await withId(new Location(), "a b").getAttributes();
    await Ajax.saveEntity(withId(new Role(), "a b"), "/role/");
    expect(requests).toEqual([
      "/user/a%20b",
      "/location/a%20b/attribute",
      "/role/a%20b",
    ]);
  });
});
