import { api } from '@/services/api';
import type { LoginResponse } from './types';

export async function loginRequest(email: string, password: string) {
  const res = await api.post<LoginResponse>('/login', { email, password });
  return res.data;
}

export async function refreshRequest() {
  const res = await api.post<LoginResponse>('/refresh');
  return res.data;
}
