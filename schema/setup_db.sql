ALTER DATABASE meow_db SET timezone TO 'UTC';

CREATE schema audit;

REVOKE USAGE ON SCHEMA audit FROM PUBLIC;
REVOKE CREATE ON schema audit FROM public;
REVOKE USAGE ON SCHEMA public FROM PUBLIC;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

GRANT USAGE ON SCHEMA audit to meow_admin;
GRANT CREATE ON SCHEMA audit to meow_admin;
GRANT USAGE ON SCHEMA public to meow_admin;
GRANT CREATE ON SCHEMA public to meow_admin;

/* grant the schema access privilege to normal users. Without schema right, user will unable to see the tables. */
GRANT USAGE ON SCHEMA audit to meow_user;
GRANT USAGE ON SCHEMA public to meow_user;
GRANT USAGE ON SCHEMA audit to meow_readonly;
GRANT USAGE ON SCHEMA public to meow_readonly;
