CREATE TABLE repositories (
    id bigint PRIMARY KEY auto_increment,    
    name varchar(255),
    owner varchar(255),
    repository_url varchar(255),
    created_at  datetime,
    updated_at  datetime
);
CREATE UNIQUE INDEX repositories_repository_url_unique_idx ON repositories(repository_url);

CREATE TABLE scans (
    id bigint PRIMARY KEY auto_increment,
    repository_id bigint,    
    repository_name varchar(255),
    repository_url varchar(255),
    findings text,
    status varchar(255),
    queued_at datetime,
    scanning_at datetime,
    finished_at datetime,
    created_at  datetime,
    updated_at  datetime
);
CREATE INDEX scans_repository_id_idx ON scans(repository_id, repository_name);
CREATE INDEX scans_status_idx ON scans(status);