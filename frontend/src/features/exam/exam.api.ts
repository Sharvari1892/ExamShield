import { api } from '@/services/api';
import type { ExamStartResponse } from './types';

export async function startExam() {
  const res = await api.post<ExamStartResponse>('/exam/start');
  return res.data;
}
