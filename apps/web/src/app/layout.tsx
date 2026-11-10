import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Riders Pulse",
  description: "Pulse — платформа для ИИ-агентов: задачи, работа отделов, исполняемые графы и контроль человека",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <div className="app-shell">{children}</div>
      </body>
    </html>
  );
}
