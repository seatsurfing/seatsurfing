import Ajax from "../util/Ajax";

export interface PasskeyListItem {
  id: string;
  name: string;
  createdAt: string;
  lastUsedAt: string | null;
}

export interface BeginRegistrationResponse {
  stateId: string;
  challenge: any;
}

export interface BeginLoginResponse {
  stateId: string;
  challenge: any;
}

export interface PasskeyChallengeResponse {
  requirePasskey: boolean;
  stateId: string;
  passkeyChallenge: any;
  allowTotpFallback: boolean;
}

// ---------------------------------------------------------------------------
// WebAuthn encoding helpers (no external library needed)
// ---------------------------------------------------------------------------

function base64urlToUint8Array(base64url: string): Uint8Array {
  const base64 = base64url.replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64.padEnd(
    base64.length + ((4 - (base64.length % 4)) % 4),
    "=",
  );
  const binary = window.atob(padded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

function uint8ArrayToBase64url(bytes: Uint8Array): string {
  let binary = "";
  bytes.forEach((b) => (binary += String.fromCharCode(b)));
  return window
    .btoa(binary)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

// Converts the server's PublicKeyCredentialCreationOptions (with base64url strings)
// into the format accepted by navigator.credentials.create()
export function prepareCreationOptions(
  serverOptions: any,
): PublicKeyCredentialCreationOptions {
  const opts = serverOptions.publicKey ?? serverOptions;
  return {
    ...opts,
    challenge: base64urlToUint8Array(opts.challenge),
    user: {
      ...opts.user,
      id: base64urlToUint8Array(opts.user.id),
    },
    excludeCredentials: (opts.excludeCredentials ?? []).map((c: any) => ({
      ...c,
      id: base64urlToUint8Array(c.id),
    })),
  };
}

// Converts the server's PublicKeyCredentialRequestOptions (with base64url strings)
// into the format accepted by navigator.credentials.get()
export function prepareRequestOptions(
  serverOptions: any,
): PublicKeyCredentialRequestOptions {
  const opts = serverOptions.publicKey ?? serverOptions;
  return {
    ...opts,
    challenge: base64urlToUint8Array(opts.challenge),
    allowCredentials: (opts.allowCredentials ?? []).map((c: any) => ({
      ...c,
      id: base64urlToUint8Array(c.id),
    })),
  };
}

// Serialise a PublicKeyCredential returned by navigator.credentials.create()
// into a plain JSON object that the backend expects
export function serializeAttestationResponse(
  cred: PublicKeyCredential,
): object {
  const response = cred.response as AuthenticatorAttestationResponse;
  return {
    id: cred.id,
    rawId: uint8ArrayToBase64url(new Uint8Array(cred.rawId)),
    type: cred.type,
    response: {
      clientDataJSON: uint8ArrayToBase64url(
        new Uint8Array(response.clientDataJSON),
      ),
      attestationObject: uint8ArrayToBase64url(
        new Uint8Array(response.attestationObject),
      ),
      transports: response.getTransports ? response.getTransports() : [],
    },
  };
}

// Serialise a PublicKeyCredential returned by navigator.credentials.get()
// into a plain JSON object that the backend expects
export function serializeAssertionResponse(cred: PublicKeyCredential): object {
  const response = cred.response as AuthenticatorAssertionResponse;
  return {
    id: cred.id,
    rawId: uint8ArrayToBase64url(new Uint8Array(cred.rawId)),
    type: cred.type,
    response: {
      clientDataJSON: uint8ArrayToBase64url(
        new Uint8Array(response.clientDataJSON),
      ),
      authenticatorData: uint8ArrayToBase64url(
        new Uint8Array(response.authenticatorData),
      ),
      signature: uint8ArrayToBase64url(new Uint8Array(response.signature)),
      userHandle: response.userHandle
        ? uint8ArrayToBase64url(new Uint8Array(response.userHandle))
        : null,
    },
  };
}

// ---------------------------------------------------------------------------
// API helpers
// ---------------------------------------------------------------------------

export default class Passkey {
  /**
   * Quick synchronous check: is the WebAuthn API present in this browser?
   * Use isPlatformAuthAvailable() for a more accurate asynchronous check
   * that also verifies a user-verifying platform authenticator is enrolled.
   */
  static isSupported(): boolean {
    return typeof window !== "undefined" && !!window.PublicKeyCredential;
  }

  /**
   * Asynchronous check that returns true only when the WebAuthn API is
   * present AND a user-verifying platform authenticator (Touch ID, Face ID,
   * Windows Hello, etc.) is available (Finding #14).
   */
  static async isPlatformAuthAvailable(): Promise<boolean> {
    if (!Passkey.isSupported()) return false;
    try {
      return await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
    } catch {
      return false;
    }
  }

  static async list(): Promise<PasskeyListItem[]> {
    return Ajax.get("/user/passkey/").then(
      (result) => result.json as PasskeyListItem[],
    );
  }

  static async beginRegistration(): Promise<BeginRegistrationResponse> {
    return Ajax.postData("/user/passkey/registration/begin", null).then(
      (result) => result.json as BeginRegistrationResponse,
    );
  }

  static async finishRegistration(
    stateId: string,
    name: string,
    credential: object,
  ): Promise<PasskeyListItem> {
    const payload = { stateId, name, credential };
    return Ajax.postData("/user/passkey/registration/finish", payload).then(
      (result) => result.json as PasskeyListItem,
    );
  }

  static async rename(id: string, name: string): Promise<void> {
    return Ajax.putData("/user/passkey/" + id, { name }).then(() => undefined);
  }

  static async deletePasskey(id: string): Promise<void> {
    return Ajax.delete("/user/passkey/" + id).then(() => undefined);
  }

  static async beginLogin(organizationId: string): Promise<BeginLoginResponse> {
    return Ajax.postData("/auth/passkey/login/begin", { organizationId }).then(
      (result) => result.json as BeginLoginResponse,
    );
  }

  static async finishLogin(
    stateId: string,
    credential: object,
  ): Promise<{ accessToken: string; refreshToken: string }> {
    const payload = { stateId, credential };
    return Ajax.postData("/auth/passkey/login/finish", payload).then(
      (result) => result.json,
    );
  }
}
