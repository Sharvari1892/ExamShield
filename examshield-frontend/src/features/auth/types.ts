export type User = {
  id: string;
  email: string;
  role: string;
};

export type LoginResponse = {
  access_token: string;
  user: User;
};