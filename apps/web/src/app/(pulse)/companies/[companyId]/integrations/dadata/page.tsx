import { ClientRedirectPage } from "../../../../../../features/navigation/ui/ClientRedirectPage";

export default function CompanyDaDataRedirect() {
  return (
    <ClientRedirectPage
      label="Переход к настройке DaData"
      to="/project-integrations/new?provider=dadata.party"
    />
  );
}
