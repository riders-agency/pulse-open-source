export const apiBase =
  (typeof process !== "undefined" && process.env.NEXT_PUBLIC_RP_API_URL
    ? process.env.NEXT_PUBLIC_RP_API_URL
    : ""
  ).replace(/\/+$/, "");
