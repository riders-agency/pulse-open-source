export type VisitorViewportBlock = {
  path?: string;
  selector?: string;
  tag?: string;
  role?: string;
  position?: string;
  is_page_chrome?: boolean;
  reason?: string;
  heading?: string;
  text?: string;
  visible_ratio?: number;
  rect?: { x?: number; y?: number; width?: number; height?: number };
};

export type VisitorViewportContext = {
  version?: number;
  captured_at?: string;
  url?: string;
  title?: string;
  viewport?: { width?: number; height?: number; scroll_x?: number; scroll_y?: number };
  primary_block?: VisitorViewportBlock | null;
  blocks?: VisitorViewportBlock[];
  visible_text?: string;
};

export type SiteChatMirrorScrollContainer = {
  path?: string;
  selector?: string;
  top?: number;
  left?: number;
  width?: number;
  height?: number;
  scroll_width?: number;
  scroll_height?: number;
};

export type SiteChatMirrorSnapshot = {
  version?: number;
  reason?: string;
  base_url?: string;
  captured_at?: string;
  html?: string;
  dom_size?: number;
  truncated?: boolean;
  scroll_containers?: SiteChatMirrorScrollContainer[];
};
