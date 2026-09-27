import React from "react";
import Ajax from "@/util/Ajax";

interface Props {
  id: string;
  src: string;
  tagName: string;
  style?: React.CSSProperties;
  onNavigate?: (path: string) => void;
  onSkipWelcomeScreen?: (reload: boolean) => void;
  // Called when the mounted element dispatches "plugin-data-changed", i.e.
  // booking data changed as a side effect of using the integration (e.g. a
  // chat assistant created or cancelled a booking).
  onDataChanged?: () => void;
}

// Plugin JS modules load asynchronously (<script type="module">), so the
// custom element they register isn't defined yet at the moment React would
// otherwise render <tag-name>. Rendering it early creates an "undefined"
// custom element instance; any property set on it before customElements.define()
// runs becomes a plain own-property that permanently shadows the class's
// accessor once upgraded (a well-known custom-elements gotcha), so e.g. the
// accessToken setter never fires again. To avoid that entirely, don't render
// the tag until its module has actually finished loading.
const modulePromises = new Map<string, Promise<void>>();

function ensurePluginModuleLoaded(src: string): Promise<void> {
  let promise = modulePromises.get(src);
  if (!promise) {
    promise = new Promise<void>((resolve, reject) => {
      const script = document.createElement("script");
      script.type = "module";
      script.src = src;
      script.onload = () => resolve();
      script.onerror = () =>
        reject(new Error(`Failed to load plugin module: ${src}`));
      document.head.appendChild(script);
    });
    modulePromises.set(src, promise);
  }
  return promise;
}

interface State {
  ready: boolean;
  error: string | null;
}

// Embeds a plugin-provided admin UI: the plugin ships a custom element (Web
// Component) that is mounted directly into the host DOM.
export default class PluginEmbed extends React.Component<Props, State> {
  private element: HTMLElement | null = null;
  private mounted = false;

  constructor(props: Props) {
    super(props);
    this.state = { ready: false, error: null };
  }

  componentDidMount() {
    this.mounted = true;
    this.setup();
    window.addEventListener(
      Ajax.ACCESS_TOKEN_REFRESHED_EVENT,
      this.handleAccessTokenRefreshed,
    );
  }

  componentDidUpdate(prevProps: Props) {
    if (
      prevProps.src !== this.props.src ||
      prevProps.tagName !== this.props.tagName
    ) {
      this.setState({ ready: false, error: null });
      this.setup();
    }
  }

  componentWillUnmount() {
    this.mounted = false;
    this.detachListeners(this.element);
    window.removeEventListener(
      Ajax.ACCESS_TOKEN_REFRESHED_EVENT,
      this.handleAccessTokenRefreshed,
    );
  }

  private setup() {
    ensurePluginModuleLoaded(this.props.src)
      .then(() => {
        if (this.mounted) {
          this.setState({ ready: true });
        }
      })
      .catch((err) => {
        console.error(err);
        if (this.mounted) {
          this.setState({
            error: err instanceof Error ? err.message : String(err),
          });
        }
      });
  }

  private handleNavigate = (e: Event) => {
    const detail = (e as CustomEvent).detail;
    if (this.props.onNavigate && detail && typeof detail.path === "string") {
      this.props.onNavigate(detail.path);
    }
  };

  private handleSkipWelcomeScreen = (e: Event) => {
    const detail = (e as CustomEvent).detail;
    if (this.props.onSkipWelcomeScreen) {
      this.props.onSkipWelcomeScreen(detail.reload);
    }
  };

  private handleDataChanged = () => {
    this.props.onDataChanged?.();
  };

  // Pull side of keeping the mounted element's access token fresh: the
  // element dispatches this when the backend rejects its current token. This
  // reuses the same refresh-token flow (and mutex) as the host's own
  // requests, so it never races a refresh already in flight. If the refresh
  // itself fails, the token is truly gone, so fall back to the host's normal
  // session-expired handling instead of leaving the panel stuck.
  private handleTokenExpired = async () => {
    const refreshToken = Ajax.PERSISTER.readRefreshTokenFromLocalStorage();
    if (!refreshToken) {
      Ajax.onUnauthorized?.();
      return;
    }
    try {
      await Ajax.refreshAccessToken(refreshToken);
    } catch {
      Ajax.onUnauthorized?.();
      return;
    }
    const credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
    if (!credentials.accessToken) {
      Ajax.onUnauthorized?.();
      return;
    }
    if (this.element) {
      (this.element as any).accessToken = credentials.accessToken;
    }
  };

  // Push side of keeping the mounted element's access token fresh: whenever
  // the host refreshes its own token (e.g. ahead of one of its own
  // requests), it broadcasts the new value so a long-lived panel never has
  // to wait for its token to actually expire first.
  private handleAccessTokenRefreshed = (e: Event) => {
    const detail = (e as CustomEvent).detail;
    if (this.element && detail && typeof detail.accessToken === "string") {
      (this.element as any).accessToken = detail.accessToken;
    }
  };

  private detachListeners(el: HTMLElement | null) {
    if (!el) {
      return;
    }
    el.removeEventListener("plugin-navigate", this.handleNavigate);
    el.removeEventListener(
      "plugin-skip-welcome-screen",
      this.handleSkipWelcomeScreen,
    );
    el.removeEventListener("plugin-data-changed", this.handleDataChanged);
    el.removeEventListener("plugin-token-expired", this.handleTokenExpired);
  }

  private attachElement = (el: HTMLElement | null) => {
    if (this.element === el) {
      return;
    }
    this.detachListeners(this.element);
    this.element = el;
    if (!el) {
      return;
    }
    const credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
    (el as any).accessToken = credentials.accessToken;
    (el as any).backendUrl = Ajax.getBackendUrl();
    el.addEventListener("plugin-navigate", this.handleNavigate);
    el.addEventListener(
      "plugin-skip-welcome-screen",
      this.handleSkipWelcomeScreen,
    );
    el.addEventListener("plugin-data-changed", this.handleDataChanged);
    el.addEventListener("plugin-token-expired", this.handleTokenExpired);
  };

  render() {
    const { id, tagName, style } = this.props;
    if (this.state.error) {
      return (
        <div className="alert alert-danger" role="alert">
          Failed to load plugin UI ({this.state.error}).
        </div>
      );
    }
    if (!this.state.ready) {
      return null;
    }
    return React.createElement(tagName, {
      id,
      ref: this.attachElement,
      style,
    });
  }
}
