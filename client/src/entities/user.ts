export type User = {
  id?: number;
  name?: string;
  last_name?: string;
  login?: string;
  email?: string;
  phone?: string;
};

export type AuthResponse = {
  token?: string;
  access_token?: string;
  user?: User;
};
