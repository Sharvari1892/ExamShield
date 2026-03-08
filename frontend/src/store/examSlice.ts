import { create } from 'zustand';
import type { Question, ExamStartResponse } from '@/features/exam/types';

type ExamState = {
  sessionId: string | null;
  questions: Question[];
  serverEndTime: number | null;
  serverOffset: number;
  setExam: (data: ExamStartResponse) => void;
};

export const useExamStore = create<ExamState>((set) => ({
  sessionId: null,
  questions: [],
  serverEndTime: null,
  serverOffset: 0,

  setExam: (data) => {
    const offset = data.server_time - Date.now();
    set({
      sessionId: data.sessionId,
      questions: data.questions,
      serverEndTime: data.end_time,
      serverOffset: offset,
    });
  },
}));
