import { useRouter } from "next/router";
import Head from "next/head";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "next-export-i18n";
import Ajax from "@/util/Ajax";
import Formatting from "@/util/Formatting";

interface KioskBooking {
  id: string;
  subject: string;
  owner: string;
  ownerVisible: boolean;
  enter: string;
  leave: string;
}

interface KioskData {
  spaceId: string;
  spaceName: string;
  locationId: string;
  locationName: string;
  timezone: string;
  status: "available" | "occupied";
  currentBooking: KioskBooking | null;
  nextBooking: KioskBooking | null;
  refreshedAt: string;
}

const KIOSK_SECRET_KEY = "kioskSecret";
const REFRESH_INTERVAL_MS = 60 * 1000;

const formatTime = (iso: string, timezone: string): string => {
  try {
    return new Date(iso).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
      timeZone: timezone,
    });
  } catch {
    return iso;
  }
};

export default function KioskPage() {
  const router = useRouter();

  // Pre-seed localStorage synchronously from ?lang= before the i18n library
  // initializes its hook state — this ensures the first effect run picks up
  // the correct language without needing a page reload.
  useState(() => {
    if (typeof window === "undefined") return;
    try {
      const lang = new URLSearchParams(window.location.search).get("lang");
      if (lang) localStorage.setItem("next-export-i18n-lang", lang);
    } catch {
      // ignore
    }
  });

  const { t } = useTranslation();

  const [data, setData] = useState<KioskData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [lastRefresh, setLastRefresh] = useState<Date | null>(null);

  const spaceId = router.query.id as string | undefined;
  const variant =
    (router.query.variant as string | undefined) === "mono" ? "mono" : "color";

  // Apply ?lang= URL param — dispatch the library's event so already-mounted
  // hook instances also update (runs once on mount, uses window.location directly).
  useEffect(() => {
    try {
      const lang = new URLSearchParams(window.location.search).get("lang");
      if (lang) {
        localStorage.setItem("next-export-i18n-lang", lang);
        document.dispatchEvent(new Event("localStorageLangChange"));
      }
    } catch {
      // ignore
    }
  }, []);

  // Store kiosk secret from URL param into localStorage, then strip from URL
  useEffect(() => {
    if (!router.isReady) return;
    const secretFromUrl = router.query.secret as string | undefined;
    if (secretFromUrl) {
      try {
        localStorage.setItem(
          KIOSK_SECRET_KEY + "_" + (spaceId ?? ""),
          secretFromUrl,
        );
      } catch {
        // localStorage may not be available
      }
      // Remove secret from URL without reloading
      const { secret, ...rest } = router.query;
      void secret; // intentionally consumed
      router.replace({ pathname: router.pathname, query: rest }, undefined, {
        shallow: true,
      });
    }
  }, [router.isReady, router.query.secret]);

  const getSecret = useCallback((): string => {
    try {
      return (
        localStorage.getItem(KIOSK_SECRET_KEY + "_" + (spaceId ?? "")) ?? ""
      );
    } catch {
      return "";
    }
  }, [spaceId]);

  const fetchData = useCallback(async () => {
    if (!spaceId) return;
    const secret = getSecret();
    if (!secret) {
      setError(t("kioskSecretMissing"));
      return;
    }
    const backendUrl = Ajax.URL ? Ajax.URL : "";
    const url = `${backendUrl}/kiosk/${spaceId}/status`;
    try {
      const res = await fetch(url, {
        headers: { Authorization: "Bearer " + secret },
      });
      if (res.status === 401) {
        setError(t("kioskAuthError"));
        return;
      }
      if (res.status === 404) {
        setError(t("kioskNotFound"));
        return;
      }
      if (!res.ok) {
        setError(t("kioskFetchError"));
        return;
      }
      const json: KioskData = await res.json();
      setData(json);
      setLastRefresh(new Date());
      setError(null);
    } catch {
      setError(t("kioskFetchError"));
    }
  }, [spaceId, getSecret, t]);

  // Initial load and periodic refresh
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const fetchDataRef = useRef(fetchData);
  useEffect(() => {
    fetchDataRef.current = fetchData;
  }, [fetchData]);
  useEffect(() => {
    if (!router.isReady || !spaceId) return;
    fetchDataRef.current();
    intervalRef.current = setInterval(
      () => fetchDataRef.current(),
      REFRESH_INTERVAL_MS,
    );
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [router.isReady, spaceId]);

  const isMono = variant === "mono";
  const isOccupied = data?.status === "occupied";

  const colorClass = isOccupied ? "kiosk-occupied" : "kiosk-available";

  return (
    <>
      <Head>
        <title>{data?.spaceName ?? t("kioskMode")}</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
      </Head>
      <style>{`
        :root { --kiosk-available: #1a7f37; --kiosk-occupied: #c0392b; --kiosk-text: #fff; }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: system-ui, sans-serif; }
        .kiosk-root {
          min-height: 100vh; display: flex; flex-direction: column;
          align-items: center; justify-content: center;
          padding: 2rem; color: var(--kiosk-text);
        }
        .kiosk-available { background: var(--kiosk-available); }
        .kiosk-occupied { background: var(--kiosk-occupied); }
        .kiosk-mono {
          background: #fff; color: #000;
          border: 8px solid #000;
        }
        .kiosk-location { font-size: 1.1rem; opacity: 0.85; margin-bottom: 0.25rem; }
        .kiosk-space { font-size: 2.5rem; font-weight: 700; margin-bottom: 1rem; }
        .kiosk-status {
          font-size: 3.5rem; font-weight: 900; letter-spacing: 0.05em;
          text-transform: uppercase; margin-bottom: 2rem;
        }
        .kiosk-card {
          width: 100%; max-width: 520px;
          background: rgba(0,0,0,0.18); border-radius: 1rem; padding: 1.25rem 1.5rem;
          margin-bottom: 1rem;
        }
        .kiosk-mono .kiosk-card {
          background: #f0f0f0; border: 2px solid #000; border-radius: 0;
          color: #000;
        }
        .kiosk-card-label {
          font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.08em;
          opacity: 0.75; margin-bottom: 0.4rem;
        }
        .kiosk-card-time { font-size: 1.75rem; font-weight: 700; }
        .kiosk-card-subject { font-size: 1.1rem; margin-top: 0.25rem; }
        .kiosk-card-owner { font-size: 0.9rem; opacity: 0.8; margin-top: 0.25rem; }
        .kiosk-footer {
          margin-top: 2rem; font-size: 0.75rem; opacity: 0.65; text-align: center;
        }
        .kiosk-error {
          min-height: 100vh; display: flex; align-items: center; justify-content: center;
          background: #fff; color: #000;
          font-size: 1rem; font-weight: 400; text-align: center; padding: 2rem;
        }
        .kiosk-mono .kiosk-status { font-size: 4rem; }
      `}</style>

      {error ? (
        <div className="kiosk-error">{error}</div>
      ) : (
        <div className={`kiosk-root ${isMono ? "kiosk-mono" : colorClass}`}>
          {data ? (
            <>
              <div className="kiosk-location">{data.locationName}</div>
              <div className="kiosk-space">{data.spaceName}</div>
              <div className="kiosk-status">
                {isOccupied ? t("kioskOccupied") : t("kioskAvailable")}
              </div>

              {data.currentBooking && (
                <div className="kiosk-card">
                  <div className="kiosk-card-label">{t("kioskNow")}</div>
                  <div className="kiosk-card-time">
                    {formatTime(data.currentBooking.enter, data.timezone)}
                    &nbsp;&ndash;&nbsp;
                    {formatTime(data.currentBooking.leave, data.timezone)}
                  </div>
                  {data.currentBooking.subject && (
                    <div className="kiosk-card-subject">
                      {data.currentBooking.subject}
                    </div>
                  )}
                  {data.currentBooking.ownerVisible &&
                    data.currentBooking.owner && (
                      <div className="kiosk-card-owner">
                        {data.currentBooking.owner}
                      </div>
                    )}
                </div>
              )}

              {data.nextBooking && (
                <div className="kiosk-card">
                  <div className="kiosk-card-label">
                    {t("kioskNext")} (
                    {Formatting.getDateOffsetText(
                      new Date(data.nextBooking.enter),
                      new Date(data.nextBooking.leave),
                    )}
                    )
                  </div>
                  <div className="kiosk-card-time">
                    {formatTime(data.nextBooking.enter, data.timezone)}
                    &nbsp;&ndash;&nbsp;
                    {formatTime(data.nextBooking.leave, data.timezone)}
                  </div>
                  {data.nextBooking.subject && (
                    <div className="kiosk-card-subject">
                      {data.nextBooking.subject}
                    </div>
                  )}
                  {data.nextBooking.ownerVisible && data.nextBooking.owner && (
                    <div className="kiosk-card-owner">
                      {data.nextBooking.owner}
                    </div>
                  )}
                </div>
              )}

              {lastRefresh && !isMono && (
                <div className="kiosk-footer">
                  {t("kioskLastRefreshed")}: {lastRefresh.toLocaleTimeString()}
                </div>
              )}
            </>
          ) : null}
        </div>
      )}
    </>
  );
}
