-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	"name" text NULL,
	hash text NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id)
);

CREATE TABLE balances (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	user_id int8 NULL,
	balance integer NULL,
	withdrawn integer NULL,
	CONSTRAINT balances_pkey PRIMARY KEY (id),
	CONSTRAINT fk_balances_user FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE TABLE operations (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	user_id int8 NULL,
	order_number text NULL,
	value integer NULL,
	operation_type text NULL,
	CONSTRAINT operations_pkey PRIMARY KEY (id),
	CONSTRAINT fk_users_operation FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE TABLE orders (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	user_id int8 NULL,
	"name" text NULL,
	hash text NULL,
	"number" text NULL,
	accrual integer NULL,
	status text NULL,
	CONSTRAINT orders_pkey PRIMARY KEY (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE balances;
DROP TABLE operations;
DROP TABLE orders;
DROP TABLE users;
-- +goose StatementEnd