CREATE TABLE repositories (
    id bigint PRIMARY KEY auto_increment,    
    name varchar(255),
    owner varchar(255),
    repository_url text,
    created_at  datetime,
    updated_at  datetime
)

CREATE TABLE scans (
    id bigint PRIMARY KEY auto_increment,
    repository_id bigint,    
    repository_name varchar(255),
    repository_url text,
    findings text,
    status varchar(255),
    queued_at datetime,
    scanning_at datetime,
    finished_at datetime,
    created_at  datetime,
    updated_at  datetime
);
