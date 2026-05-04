CREATE TABLE IF NOT EXISTS vault_meta (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    vault_name TEXT NOT NULL,
    kdf_algo TEXT NOT NULL,
    kdf_salt BLOB NOT NULL,
    kdf_time_cost INTEGER NOT NULL,
    kdf_memory_cost INTEGER NOT NULL,
    kdf_parallelism INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS vault_key_check (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    nonce BLOB NOT NULL,
    cipher_text BLOB NOT NULL,
    created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    username_enc BLOB,
    password_enc BLOB,
    website_enc BLOB,
    notes_enc BLOB,
    nonce_username BLOB,
    nonce_password BLOB,
    nonce_website BLOB,
    nonce_notes BLOB,
    category TEXT NOT NULL DEFAULT 'login',
    favorite INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS item_tags (
    item_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (item_id, tag_id),
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
);
