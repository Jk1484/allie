CREATE TABLE mages (
    id serial primary key,
    username varchar(100) unique,
    password varchar(100),
    hp int not null default 100
);
