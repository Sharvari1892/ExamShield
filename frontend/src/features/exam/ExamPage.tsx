import { useEffect, useState } from 'react';
import { useExamStore } from '@/store/examSlice';
import { startExam } from './exam.api';
import { dbPromise } from '@/offline/db';
import type { Question } from './types';

export function ExamPage() {
  const { questions, serverEndTime, serverOffset, setExam } = useExamStore();
  const [remainingSeconds, setRemainingSeconds] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function bootstrap() {
      try {
        const data = await startExam();
        setExam(data);

        const db = await dbPromise;
        await db.put('session', data, 'active');
        setLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to start exam session');
        setLoading(false);
      }
    }
    bootstrap();
  }, [setExam]);

  useEffect(() => {
    if (serverEndTime === null) return;

    const interval = setInterval(() => {
      const remaining = serverEndTime - (Date.now() + serverOffset);
      setRemainingSeconds(Math.max(0, Math.floor(remaining / 1000)));
    }, 1000);

    return () => clearInterval(interval);
  }, [serverEndTime, serverOffset]);

  if (loading) {
    return (
      <div style={{ padding: '20px' }}>
        <h2>Loading exam session...</h2>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '20px', color: 'red' }}>
        <h2>Error</h2>
        <p>{error}</p>
      </div>
    );
  }

  return (
    <div style={{ padding: '20px' }}>
      <h2>Exam Session</h2>
      <div style={{ fontSize: '18px', fontWeight: 'bold', marginBottom: '20px' }}>
        Time Remaining: {remainingSeconds}s
      </div>
      <div>
        {questions.length === 0 ? (
          <p>No questions available</p>
        ) : (
          questions.map((q: Question, idx: number) => (
            <div key={q.id || idx} style={{ marginBottom: '15px' }}>
              <strong>Q{idx + 1}:</strong> {q.content || q.text || 'No content'}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
