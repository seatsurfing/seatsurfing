import Ajax from "@/util/Ajax";

export function getIcal(bookingId: string) {
  const credentials = Ajax.PERSISTER.readCredentialsFromSessionStorage();
  let options: RequestInit = Ajax.getFetchOptions(
    "GET",
    credentials.accessToken,
    null,
  );
  fetch(Ajax.getBackendUrl() + "/booking/" + bookingId + "/ical", options).then(
    (response) => {
      if (!response.ok) {
        return;
      }
      let contentDisposition = response.headers.get("Content-Disposition");
      let filename = "seatsurfing.ics";
      if (contentDisposition) {
        let filenameMatch = contentDisposition.match(/filename="(.+)"/);
        if (filenameMatch && filenameMatch[1]) {
          filename = filenameMatch[1];
        }
      }
      response
        .blob()
        .then((data) => {
          const blob = new Blob([data], { type: "text/calendar" });
          const url = window.URL.createObjectURL(blob);
          let a = document.createElement("a");
          a.style = "display: none";
          a.href = url;
          a.download = filename;
          document.body.appendChild(a);
          a.click();
          window.URL.revokeObjectURL(url);
        })
        .catch(() => {});
    },
  );
}
