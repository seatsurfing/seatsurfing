export default class JwtDecoder {
  static getPayload(jwt: string): any {
    let tokens = jwt.split(".");
    if (tokens.length != 3) {
      return null;
    }
    let payload = "{}";
    if (typeof window !== "undefined") {
      payload = window.atob(tokens[1]);
    }
    let json = JSON.parse(payload);
    return json;
  }

  static getExpiryDate(jwt: string): Date {
    const payload = JwtDecoder.getPayload(jwt);
    if (!payload || !payload.exp) {
      // Fallback to 5 minutes from now if exp claim is missing
      return new Date(new Date().getTime() + 5 * 60 * 1000);
    }
    // JWT exp is in seconds, convert to milliseconds
    return new Date(payload.exp * 1000);
  }
}
