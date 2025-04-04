Create TABLE api_key (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE,
    prefix CHAR(8) NOT NULL,
    hashed_key CHAR(60) NOT NULL,
    name VARCHAR(150) NOT NULL,
    expires_at TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    UNIQUE (user_uuid, name)
);
