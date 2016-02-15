
GRANT SELECT, INSERT, UPDATE, DELETE, REFERENCES ON TABLE users                 to meow_user;
GRANT SELECT, INSERT, UPDATE, DELETE, REFERENCES ON TABLE cats                  to meow_user;

GRANT SELECT ON TABLE users                 to meow_readonly;
GRANT SELECT ON TABLE cats                  to meow_readonly;
