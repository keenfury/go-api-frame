create table user (
    id serial,
    first_name varchar(50),
    age int not null,
    active boolean not null default true,
    primary key(id)
);

create table client (
    id serial,
    first_name varchar(50),
    last_name varchar(80),
    phone_number varchar(20) not null,
    age int not null,
    active boolean not null default true,
    primary key(id)
);

create table customer (
    id serial,
    first_name varchar(100) not null,
    primary key(id)
);