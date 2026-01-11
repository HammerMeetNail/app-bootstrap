DROP TRIGGER IF EXISTS update_notes_updated_at ON notes;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS magic_link_tokens;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
