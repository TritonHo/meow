
insert into users(id, email, password_digest, first_name, last_name)
values
('eeee1df4-9fae-4e32-98c1-88f850a00001', 'Susan.Wong@abc.com', '$2a$10$Ba/oRmxnRx0D5/dZlMGqs.4rF2NC.pKbouzUKTHFaZ.we.1YFz5cO', 'Susan', 'Wong');



insert into cats(id, user_id, name, gender) 
values
('ffff1df4-9fae-4e32-98c1-88f850a00001', 'eeee1df4-9fae-4e32-98c1-88f850a00001', 'LittleWhite', 'FEMALE');


