import type { Metadata } from "next";
import { TechShowcase } from "../../../features/tech/TechShowcase";

export const metadata: Metadata = {
  title: "Pulse — платформа для ИИ-агентов | Riders Agency",
  description: "Попросите в чате — агенты подготовят задачу, разберут письмо, обновят учебный сайт и подготовят счёт. Записи работы, исполняемые графы и контроль человека.",
};

export default function TechPage() {
  return <TechShowcase />;
}
