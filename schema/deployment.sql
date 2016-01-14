/*
	FIXME: it does not create user with just enough privilege
*/


/* 
	should be run by postgres super user
	For example, run the following command in the database server: sudo -u postgres psql
*/
CREATE ROLE demo_admin LOGIN PASSWORD 'password' NOSUPERUSER INHERIT NOCREATEDB NOCREATEROLE NOREPLICATION;

CREATE DATABASE demo_db with ENCODING = 'UTF8' TABLESPACE = pg_default LC_COLLATE = 'en_US.UTF-8' LC_CTYPE = 'en_US.UTF-8' CONNECTION LIMIT = -1 template=template0;

alter DATABASE demo_db OWNER TO demo_admin

ALTER DATABASE demo_db SET timezone TO 'UTC';

/* 
	should be run by postgres, and inside demo_db
 	For example, run the following command in the database server: sudo -u postgres psql demo_db
*/
CREATE EXTENSION "uuid-ossp";


