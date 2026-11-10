export type Company = {
  company_id: string;
  name: string;
  slug?: string;
  role?: string;
};
export type Project = {
  project_id: string;
  company_id?: string;
  name: string;
  slug: string;
  hook_key?: string;
  hook_secret?: string;
};

export type WorkspaceContext = {
  user_id: string;
  active_company_id: string;
  active_project_id: string;
};
