-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- スケジュールテーブル
CREATE TABLE IF NOT EXISTS schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    description TEXT,
    location TEXT,
    owner_id INTEGER NOT NULL, -- このスケジュールが属するカレンダーの所有者
    creator_id INTEGER NOT NULL, -- このスケジュールを作成したユーザー
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (creator_id) REFERENCES users(id)
);

-- スケジュール参加者テーブル (多対多)
CREATE TABLE IF NOT EXISTS schedule_participants (
    schedule_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    PRIMARY KEY (schedule_id, user_id),
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- スケジュール更新日時のトリガー
CREATE TRIGGER IF NOT EXISTS update_schedules_updated_at
AFTER UPDATE ON schedules
FOR EACH ROW
BEGIN
    UPDATE schedules SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;