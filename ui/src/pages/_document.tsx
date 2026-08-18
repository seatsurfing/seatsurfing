import Document, {
  Html,
  Head,
  Main,
  NextScript,
  DocumentProps,
} from "next/document";
import { randomBytes } from "crypto";
import RuntimeConfig from "@/components/RuntimeConfig";

type Props = DocumentProps & {
  // add custom document props
};

class Doc extends Document<Props> {
  render() {
    const nonce = randomBytes(128).toString("base64");
    const csp = new Map<string, string[]>();
    csp.set("default-src", ["'self'"]);
    csp.set("form-action", ["'self'"]);
    csp.set("img-src", ["'self'", "data:", "https:"]);
    csp.set("style-src", ["'self'", "data:", "'unsafe-inline'"]);
    csp.set("object-src", ["data:"]);
    csp.set("base-uri", ["'none'"]);
    csp.set("script-src", [
      "'self'",
      "'nonce-" + nonce + "'",
      "'strict-dynamic'",
    ]);
    if (process.env.NODE_ENV.toLowerCase() === "development") {
      csp.set("connect-src", ["'self'", "http://localhost:8080"]);
      csp.set(
        "script-src",
        Object.assign(
          [],
          csp.get("script-src")?.concat(["'unsafe-eval'", "'unsafe-inline'"]),
        ),
      );
    }
    let cspString = "";
    csp.keys().forEach((key) => {
      cspString += `${key} ${csp.get(key)?.join(" ")}; `;
    });
    return (
      <Html lang={RuntimeConfig.getLanguage()}>
        <Head nonce={nonce}>
          <meta name="robots" content="noindex" />
          <meta httpEquiv="Content-Security-Policy" content={cspString} />
          {/* Sets data-bs-theme before first paint to avoid a light/dark flash; admin pages manage theme separately */}
          <script
            nonce={nonce}
            dangerouslySetInnerHTML={{
              __html: `(function(){try{if(/\\/admin(\\/|$)/.test(window.location.pathname))return;var m=window.localStorage.getItem("theme")||"auto";var d=m==="dark"||(m==="auto"&&window.matchMedia("(prefers-color-scheme: dark)").matches);document.documentElement.setAttribute("data-bs-theme",d?"dark":"light");}catch(e){}})();`,
            }}
          />
        </Head>
        <body>
          <Main />
          <NextScript nonce={nonce} />
        </body>
      </Html>
    );
  }
}

export default Doc;
