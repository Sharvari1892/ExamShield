CREATE TABLE exams (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    duration_seconds INT NOT NULL,
    published BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE questions (
    id UUID PRIMARY KEY,
    exam_id UUID REFERENCES exams(id) ON DELETE CASCADE,
    difficulty_level INT NOT NULL,
    content TEXT NOT NULL,
    options JSONB,
    correct_answer TEXT,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE exam_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    exam_id UUID REFERENCES exams(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status TEXT NOT NULL,
    grace_used_seconds INT DEFAULT 0,
    device_fingerprint TEXT,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE answers (
    id UUID PRIMARY KEY,
    session_id UUID REFERENCES exam_sessions(id) ON DELETE CASCADE,
    question_id UUID REFERENCES questions(id),
    answer_data JSONB,
    version INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMP DEFAULT now(),
    UNIQUE(session_id, question_id)
);
