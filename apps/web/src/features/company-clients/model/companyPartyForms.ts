import type { Client, DaDataPartySuggestion } from "../../../lib/api";

export type CompanyClientForm = {
  client_type: Client["client_type"];
  name: string;
  short_name: string;
  full_name: string;
  inn: string;
  kpp: string;
  ogrn: string;
  ogrnip: string;
  legal_address: string;
  actual_address: string;
  okved: string;
  comment: string;
  dadata_raw_json?: unknown;
};

export type CompanyLegalEntityForm = {
  entity_type: "individual_entrepreneur" | "legal_entity";
  name: string;
  short_name: string;
  full_name: string;
  inn: string;
  kpp: string;
  ogrn: string;
  ogrnip: string;
  legal_address: string;
  actual_address: string;
  okved: string;
  comment: string;
  dadata_raw_json?: unknown;
};

export const emptyCompanyClientForm: CompanyClientForm = {
  client_type: "legal_entity",
  name: "",
  short_name: "",
  full_name: "",
  inn: "",
  kpp: "",
  ogrn: "",
  ogrnip: "",
  legal_address: "",
  actual_address: "",
  okved: "",
  comment: "",
};

export const emptyCompanyLegalEntityForm: CompanyLegalEntityForm = {
  entity_type: "legal_entity",
  name: "",
  short_name: "",
  full_name: "",
  inn: "",
  kpp: "",
  ogrn: "",
  ogrnip: "",
  legal_address: "",
  actual_address: "",
  okved: "",
  comment: "",
};

export function companyClientFormFromDaData(item: DaDataPartySuggestion, comment: string): CompanyClientForm {
  return {
    client_type: item.client_type,
    name: item.name || item.short_name || item.full_name || "",
    short_name: item.short_name ?? "",
    full_name: item.full_name ?? "",
    inn: item.inn ?? "",
    kpp: item.kpp ?? "",
    ogrn: item.ogrn ?? "",
    ogrnip: item.ogrnip ?? "",
    legal_address: item.legal_address ?? "",
    actual_address: item.actual_address ?? item.legal_address ?? "",
    okved: item.okved ?? "",
    comment,
    dadata_raw_json: item.dadata_raw_json ?? item,
  };
}

export function legalEntityFormFromDaData(item: DaDataPartySuggestion, comment: string): CompanyLegalEntityForm {
  return {
    entity_type: item.client_type === "individual_entrepreneur" ? "individual_entrepreneur" : "legal_entity",
    name: item.name ?? "",
    short_name: item.short_name ?? "",
    full_name: item.full_name ?? "",
    inn: item.inn ?? "",
    kpp: item.kpp ?? "",
    ogrn: item.ogrn ?? "",
    ogrnip: item.ogrnip ?? "",
    legal_address: item.legal_address ?? "",
    actual_address: item.actual_address ?? item.legal_address ?? "",
    okved: item.okved ?? "",
    comment,
    dadata_raw_json: item.dadata_raw_json ?? item,
  };
}
