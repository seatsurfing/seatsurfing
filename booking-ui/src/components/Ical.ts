import { Ajax } from "seatsurfing-commons";

export function getIcal (bookingId: string) {
    let options: RequestInit = Ajax.getFetchOptions("GET", Ajax.CREDENTIALS.accessToken, null);
    fetch(Ajax.getBackendUrl() + '/booking/' + bookingId + '/ical', options).then((response) => {
      if (!response.ok) {
        return;
      }
      response.blob().then((data) => {
        const blob = new Blob([data], { type: 'text/calendar' });
        const url = window.URL.createObjectURL(blob);
        let a = document.createElement("a");
        a.style = "display: none";
        a.href = url;
        a.download = "seatsurfing.ics";
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
      }).catch(() => {});
    });
  }