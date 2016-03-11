CREATE OR REPLACE FUNCTION audit_cats_function()
returns TRIGGER AS $$
begin
	IF TG_OP = 'INSERT' then
		insert into audit.cats(
			id, action_time, 
			user_id_new, name_new, gender_new
		)
		values(
			new.id, now(), 
			new.user_id, new.name, new.gender
		);
	END IF;

	IF	TG_OP = 'UPDATE' then
		insert into audit.cats(
			id, action_time, 
			user_id_old, name_old, gender_old, 
			user_id_new, name_new, gender_new
		)
		values(
			old.id, now(), 
			old.user_id, old.name, old.gender,
			new.user_id, new.name, new.gender
		);
	END IF;
	IF TG_OP = 'DELETE' then
		insert into audit.cats(
			id, action_time, 
			user_id_old, name_old, gender_old 
		)
		values(
			old.id, now(), 
			old.user_id, old.name, old.gender
		);
	END IF;

	RETURN NULL;
end;
$$
LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER audit_cats AFTER INSERT or update or delete
ON cats FOR each row 
execute procedure audit_cats_function();

