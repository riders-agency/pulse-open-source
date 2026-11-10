"use client";

import { useAccountSettings } from "./model/useAccountSettings";
import { AccountApiKeysCard } from "./ui/AccountApiKeysCard";
import { AccountProfileCard } from "./ui/AccountProfileCard";
import { AccountSessionsCard } from "./ui/AccountSessionsCard";

export default function AccountPage() {
  const settings = useAccountSettings();

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Аккаунт</h1>
          <p className="muted">Профиль и активные устройства</p>
        </div>
      </div>

      <div className="stack-lg">
        <AccountProfileCard settings={settings} />
        <AccountSessionsCard settings={settings} />
        <AccountApiKeysCard settings={settings} />
      </div>
    </main>
  );
}
