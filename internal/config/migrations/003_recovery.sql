CREATE TABLE IF NOT EXISTS vault_recovery (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    kdf_algo TEXT NOT NULL,
    kdf_salt BLOB NOT NULL,
    kdf_time_cost INTEGER NOT NULL,
    kdf_memory_cost INTEGER NOT NULL,
    kdf_parallelism INTEGER NOT NULL,
    nonce BLOB NOT NULL,
    cipher_text BLOB NOT NULL,
    created_at DATETIME NOT NULL
);
