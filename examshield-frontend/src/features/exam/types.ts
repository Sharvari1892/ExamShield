export type Question = {
  id: string;
  content: string;
  text?: string;
  options?: string[];
};

export type ExamStartResponse = {
  sessionId: string;
  questions: Question[];
  server_time: number;
  end_time: number;
};
