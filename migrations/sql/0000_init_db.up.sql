CREATE TABLE repositories (
    id bigint PRIMARY KEY auto_increment,
    repository text,
    created_at  datetime,
    updated_at  datetime
)

CREATE TABLE scans (
    id bigint PRIMARY KEY auto_increment,
    repository_id bigint,    
    identifier text,
    findings json,
    status varchar(255),
    queued_at TIMESTAMP DEFAULT NOW(),
    created_at  datetime,
    updated_at  datetime
);

