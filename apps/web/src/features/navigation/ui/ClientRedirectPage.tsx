"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { resolveRedirectTarget } from "../model/redirectTarget";

type ClientRedirectPageProps = {
  label: string;
  to: string;
};

export function ClientRedirectPage({ label, to }: ClientRedirectPageProps) {
  const router = useRouter();

  useEffect(() => {
    const target = resolveRedirectTarget(to, window.location.href);

    if (target.kind === "blocked") return;
    if (target.kind === "external") {
      window.location.replace(target.href);
      return;
    }

    router.replace(target.href);
  }, [router, to]);

  return <main className="page" aria-label={label} />;
}
