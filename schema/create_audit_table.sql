create table audit.cats 
(
	id uuid,
	action_time timestamp with time zone not null default current_timestamp,

	user_id_old uuid,
	user_id_new uuid,

	name_old character varying(1000),
	name_new character varying(1000),
	gender_old character varying(1000),
	gender_new character varying(1000),

	CONSTRAINT "cats_audit_pk" PRIMARY KEY (id, action_time)
);
