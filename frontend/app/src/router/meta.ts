export type RouteMeta = {
  requiresAuth?: boolean;
  requiresAdmin?: boolean;
  title?: string;
  hideInMenu?: boolean;
};

export const defaultAuthMeta: RouteMeta = {
  requiresAuth: true,
  requiresAdmin: false
};
