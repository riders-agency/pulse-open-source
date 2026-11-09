import { AuthGate } from "../../features/auth/AuthGate";

export default function PulseAppLayout({ children }: { children: React.ReactNode }) {
  return <AuthGate>{children}</AuthGate>;
}
